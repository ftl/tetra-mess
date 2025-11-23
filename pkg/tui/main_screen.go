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

type RadioData struct {
	Position    data.Position
	Measurement quality.Measurement
}

type TracingStatus struct {
	Filename string
	Active   bool
}

type MainScreen struct {
	logic *Logic

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
	case *Logic:
		s.logic = msg
	case error:
		s.userMessage = fmt.Sprintf("E: %s", msg.Error())
	case string:
		s.userMessage = msg
	case RadioData:
		return s.handleRadioData(msg)
	case TracingStatus:
		return s.handleTracingStatus(msg)
	}
	s.lacTable, cmd = s.lacTable.Update(msg)
	return s, cmd
}

func (s MainScreen) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, s.keyMap.ToggleTrace):
		return s, s.logic.ToggleTrace
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
			fmt.Sprintf("%d", report.CurrentRSSI()),
			fmt.Sprintf("%d", report.CurrentGAN()),
			fmt.Sprintf("%d", report.MinRSSI),
			fmt.Sprintf("%d", report.AverageRSSI()),
			fmt.Sprintf("%d", report.MaxRSSI),
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

	ganStyle := lipgloss.NewStyle().Foreground(ganToANSIColor(s.currentGAN)).Reverse(true)
	sldStyle := lipgloss.NewStyle().Foreground(sldToANSIColor(s.currentSLD)).Reverse(true)
	serverStyle := lipgloss.NewStyle().Foreground(serversToANSIColor(s.currentServers)).Reverse(true)

	currentBox := lipgloss.JoinVertical(
		lipgloss.Left,
		headingStyle.Render("Current BS"),
		fmt.Sprintf("LAC: %d", s.currentLAC),
		ganStyle.Render(fmt.Sprintf("RSSI: %d", s.currentRSSI)),
		ganStyle.Render(fmt.Sprintf("GAN: %d", s.currentGAN)),
		fmt.Sprintf("Cx: %d", s.currentCx),
		sldStyle.Render(fmt.Sprintf("SLD: %d", s.currentSLD)),
		serverStyle.Render(fmt.Sprintf("Servers: %d", s.currentServers)),
	)

	averageBox := lipgloss.JoinVertical(
		lipgloss.Left,
		headingStyle.Render("Average"),
		fmt.Sprintf("Count: %d", s.averageCount),
		fmt.Sprintf("RSSI: %d", s.averageRSSI),
		fmt.Sprintf("GAN: %d", s.averageGAN),
		"",
		fmt.Sprintf("SLD: %d", s.averageSLD),
	)

	cellWidth := s.width / 10
	statusCell := lipgloss.NewStyle()
	statusBarBox := lipgloss.JoinHorizontal(
		lipgloss.Top,
		statusCell.Width(2*cellWidth).Render(s.device),
		" | ",
		statusCell.Width(4*cellWidth).Render(s.traceFilename),
		" | ",
		statusCell.Render(s.userMessage),
	)

	mainScreen := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			lipgloss.JoinVertical(
				lipgloss.Left,
				boxStyle.Width(30).Render(positionBox),
				boxStyle.Render(
					lipgloss.JoinHorizontal(
						lipgloss.Top,
						boxStyle.Width(15).Render(currentBox),
						boxStyle.Width(15).Render(averageBox),
					),
				),
			),
			tableStyle.MaxHeight(14).Render(s.lacTable.View()),
		),
		statusBarStyle.Width(s.width).Render(statusBarBox),
		helpStyle.Width(s.width).Render(s.help.View(s.keyMap)),
	)

	screenStyle := lipgloss.NewStyle().MaxWidth(s.width).MaxHeight(s.height)
	return screenStyle.Render(mainScreen)
}
