package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ftl/tetra-mess/pkg/data"
	"github.com/ftl/tetra-mess/pkg/quality"
)

type TracingStatus struct {
	Filename string
	Active   bool
}

type ConnectionClosed struct{}

type MainScreen struct {
	app *App

	// UI state
	version        string
	utmField       string
	latitude       float64
	longitude      float64
	satellites     int
	lastScan       string
	currentLAC     uint32
	currentRSSI    int
	currentCx      int
	currentGAN     int
	currentSLD     int
	currentServers int
	averageCount   int
	averageRSSI    int
	averageGAN     int
	averageSLD     int

	// status bar content
	userMessage   string
	device        string
	traceFilename string
	traceActive   bool

	// UI widgets
	width    int
	height   int
	keyMap   KeyMap
	help     help.Model
	lacTable table.Model

	// data
	currentPosition data.Position
	qualityReport   *quality.QualityReport
}

func NewMainScreen(version, device string) MainScreen {
	return MainScreen{
		version:         version,
		device:          device,
		currentPosition: data.NoPosition,
		qualityReport:   quality.NewQualityReport(),

		keyMap: DefaultKeyMap,
		help:   help.New(),
		lacTable: table.New(
			table.WithColumns([]table.Column{
				{Title: "LAC", Width: 5},
				{Title: "RSSI", Width: 4},
				{Title: "GAN", Width: 3},
				{Title: "Min", Width: 4},
				{Title: "Avg", Width: 4},
				{Title: "Max", Width: 4},
			}),
			table.WithStyles(table.Styles{
				Selected: tableSelectedStyle,
				Header:   tableHeaderStyle,
				Cell:     tableCellStyle,
			}),
		),
	}
}

func (s MainScreen) Init() tea.Cmd {
	return tea.SetWindowTitle("tetra-mess " + s.version)
}

func (s MainScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
	case tea.KeyMsg:
		return s.handleKey(msg)
	case *App:
		s.app = msg
	case error:
		s.userMessage = fmt.Sprintf("E: %s", msg.Error())
	case string:
		s.userMessage = msg
	case RadioData:
		return s.handleRadioData(msg)
	case TracingStatus:
		return s.handleTracingStatus(msg)
	case ConnectionClosed:
		return s, tea.Quit
	}
	s.lacTable, cmd = s.lacTable.Update(msg)
	return s, cmd
}

func (s MainScreen) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, s.keyMap.ToggleTrace):
		return s, s.app.ToggleTrace
	case key.Matches(msg, s.keyMap.Help):
		s.help.ShowAll = !s.help.ShowAll
		return s, nil
	case key.Matches(msg, s.keyMap.Quit):
		return s, tea.Quit
	default:
		return s, nil
	}
}

func (s MainScreen) handleRadioData(msg RadioData) (tea.Model, tea.Cmd) {
	s.currentPosition = msg.Position
	s.qualityReport.AddMeasurement(msg.Measurement)

	s.utmField = s.currentPosition.ToUTMField().FieldID()
	s.latitude = s.currentPosition.Latitude
	s.longitude = s.currentPosition.Longitude
	s.satellites = s.currentPosition.Satellites
	s.lastScan = s.currentPosition.Timestamp.Local().Format("02.01.2006 15:04:05")

	fieldReport := s.qualityReport.FieldReportByUTM(s.currentPosition.ToUTMField())
	bestServer := msg.Measurement.BestServer()
	s.currentLAC = bestServer.LAC
	s.currentRSSI = bestServer.RSSI
	s.currentCx = bestServer.Cx
	s.currentGAN = data.RSSIToGAN(bestServer.RSSI)
	s.currentSLD = msg.Measurement.SignalLevelDifference()
	s.currentServers = msg.Measurement.UsableServers()
	s.averageCount = len(fieldReport.Measurements)
	s.averageRSSI = fieldReport.AverageRSSI()
	s.averageGAN = fieldReport.AverageGAN()
	s.averageSLD = fieldReport.AverageSignalLevelDifference()

	lacReports := fieldReport.LACReportsByRSSI()
	rows := make([]table.Row, len(lacReports))
	for i, report := range lacReports {
		rows[i] = table.Row{
			fmt.Sprintf("%d", report.LAC),
			fmt.Sprintf("% 4d", report.CurrentRSSI()),
			fmt.Sprintf("% 3d", report.CurrentGAN()),
			fmt.Sprintf("% 3d", report.MinRSSI),
			fmt.Sprintf("% 3d", report.AverageRSSI()),
			fmt.Sprintf("% 3d", report.MaxRSSI),
		}
	}
	s.lacTable.SetRows(rows)

	return s, nil
}

func (s MainScreen) handleTracingStatus(msg TracingStatus) (tea.Model, tea.Cmd) {
	s.traceFilename = msg.Filename
	s.traceActive = msg.Active
	return s, nil
}

func (s MainScreen) View() string {
	positionBox := lipgloss.JoinVertical(
		lipgloss.Left,
		headingStyle.Render("Position"),
		s.utmField,
		fmt.Sprintf("%010.7f %010.7f", s.latitude, s.longitude),
		fmt.Sprintf("Satellites: %d", s.currentPosition.Satellites),
		fmt.Sprintf("%s", s.currentPosition.Timestamp.Local().Format("02.01.2006 15:04:05")),
	)

	positionStyle := lipgloss.NewStyle().Inherit(boxStyle).Padding(0, 1)
	if s.currentPosition.Satellites == 0 {
		positionStyle = positionStyle.Foreground(ANSIRed).Reverse(true)
	}
	ganStyle := lipgloss.NewStyle().Foreground(ganToANSIColor(s.currentGAN)).Reverse(true)
	sldStyle := lipgloss.NewStyle().Foreground(sldToANSIColor(s.currentSLD)).Reverse(true)
	serverStyle := lipgloss.NewStyle().Foreground(serversToANSIColor(s.currentServers)).Reverse(true)

	currentBox := lipgloss.JoinVertical(
		lipgloss.Left,
		headingStyle.Render("Current BS"),
		fmt.Sprintf("LAC: % 7d", s.currentLAC),
		ganStyle.Render(fmt.Sprintf("RSSI: % 6d", s.currentRSSI)),
		ganStyle.Render(fmt.Sprintf("GAN: % 7d", s.currentGAN)),
		fmt.Sprintf("Cx: % 8d", s.currentCx),
		sldStyle.Render(fmt.Sprintf("SLD: % 7d", s.currentSLD)),
		serverStyle.Render(fmt.Sprintf("Servers: %3d", s.currentServers)),
	)

	averageBox := lipgloss.JoinVertical(
		lipgloss.Left,
		headingStyle.Render("Average"),
		fmt.Sprintf("Count: % 5d", s.averageCount),
		fmt.Sprintf("RSSI: % 6d", s.averageRSSI),
		fmt.Sprintf("GAN: % 7d", s.averageGAN),
		"",
		fmt.Sprintf("SLD: % 7d", s.averageSLD),
		"",
	)

	cellWidth := (s.width - 6) / 10
	statusCell := lipgloss.NewStyle()
	statusBarBox := lipgloss.JoinHorizontal(
		lipgloss.Top,
		statusCell.Width(2*cellWidth).Render(s.device),
		" | ",
		statusCell.Width(4*cellWidth).Render(s.traceFilename),
		" | ",
		statusCell.Width(4*cellWidth).Render(s.userMessage),
	)

	mainScreen := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			lipgloss.JoinVertical(
				lipgloss.Left,
				positionStyle.Width(30).Render(positionBox),
				lipgloss.JoinHorizontal(
					lipgloss.Top,
					boxStyle.Width(14).Render(currentBox),
					boxStyle.Width(14).Render(averageBox),
				),
			),
			boxStyle.Width(38).Render(
				tableStyle.MaxHeight(14).Render(s.lacTable.View()),
			),
		),
		statusBarStyle.Width(s.width).Render(statusBarBox),
		helpStyle.Width(s.width).Render(s.help.View(s.keyMap)),
	)

	screenStyle := lipgloss.NewStyle().MaxWidth(s.width).MaxHeight(s.height)
	return screenStyle.Render(mainScreen)
}
