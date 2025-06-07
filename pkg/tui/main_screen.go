package tui

import (
	"fmt"
	"log"

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

type MainScreen struct {
	logic *Logic

	// UI state
	version     string
	utmField    string
	latitude    float64
	longitude   float64
	satellites  int
	lastScan    string
	currentLAC  uint32
	currentRSSI int
	currentCx   int
	currentGAN  int
	currentSLD  int
	averageRSSI int
	averageGAN  int
	averageSLD  int

	// UI widgets
	width    int
	height   int
	lacTable table.Model

	// data
	currentPosition data.Position
	qualityReport   *quality.QualityReport
}

func NewMainScreen(version string) MainScreen {
	return MainScreen{
		version:         version,
		currentPosition: data.NoPosition,
		qualityReport:   quality.NewQualityReport(),

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
				Selected: lipgloss.NewStyle(),
				Header:   lipgloss.NewStyle().Bold(true).Padding(0, 1),
				Cell:     lipgloss.NewStyle().Padding(0, 1),
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
		// TODO: display any error messages in the TUI
		log.Println(msg.Error())
	case string:
		// TODO: display any messages in the TUI
		log.Println(msg)
	case RadioData:
		return s.handleRadioData(msg)
	}
	s.lacTable, cmd = s.lacTable.Update(msg)
	return s, cmd
}

func (s MainScreen) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return s, tea.Quit
	case "t":
		return s, s.logic.ToggleTrace
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

func (s MainScreen) View() string {
	positionBox := lipgloss.JoinVertical(
		lipgloss.Left,
		headingStyle.Render("Position"),
		s.utmField,
		fmt.Sprintf("%010.7f %010.7f", s.latitude, s.longitude),
		fmt.Sprintf("Satellites: %d", s.currentPosition.Satellites),
		fmt.Sprintf("%s\n", s.currentPosition.Timestamp.Local().Format("02.01.2006 15:04:05")),
	)

	currentBox := lipgloss.JoinVertical(
		lipgloss.Left,
		headingStyle.Render("Current BS"),
		fmt.Sprintf("LAC: %d", s.currentLAC),
		fmt.Sprintf("RSSI: %d", s.currentRSSI),
		fmt.Sprintf("GAN: %d", s.currentGAN),
		fmt.Sprintf("Cx: %d", s.currentCx),
		fmt.Sprintf("SLD: %d", s.currentSLD),
	)

	averageBox := lipgloss.JoinVertical(
		lipgloss.Left,
		headingStyle.Render("Average"),
		"",
		fmt.Sprintf("RSSI: %d", s.averageRSSI),
		fmt.Sprintf("GAN: %d", s.averageGAN),
		"",
		fmt.Sprintf("SLD: %d", s.averageSLD),
	)

	mainScreen := lipgloss.JoinHorizontal(
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
		s.lacTable.View(),
	)

	docStyle := lipgloss.NewStyle().MaxWidth(s.width).MaxHeight(s.height)
	return docStyle.Render(mainScreen + "\n")
}
