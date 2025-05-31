package quality

import (
	"slices"

	"github.com/ftl/tetra-mess/pkg/data"
)

type QualityReport struct {
	fieldsByUTM map[string]*FieldReport
}

func NewQualityReport() *QualityReport {
	return &QualityReport{
		fieldsByUTM: make(map[string]*FieldReport),
	}
}

func (a *QualityReport) AddMeasurement(measurement Measurement) {
	for _, dataPoint := range measurement.dataPoints {
		a.Add(dataPoint)
	}
}

func (a *QualityReport) Add(dataPoint data.DataPoint) {
	fieldID := dataPoint.UTMField().FieldID()

	field, ok := a.fieldsByUTM[fieldID]
	if !ok {
		field = NewFieldReport(dataPoint.UTMField())
		a.fieldsByUTM[fieldID] = field
	}
	field.Add(dataPoint)
}

func (a *QualityReport) FieldReports() []FieldReport {
	stats := make([]FieldReport, 0, len(a.fieldsByUTM))
	for _, field := range a.fieldsByUTM {
		stats = append(stats, *field)
	}
	return stats
}

func (a *QualityReport) FieldReportByUTM(utmField data.UTMField) FieldReport {
	result, ok := a.fieldsByUTM[utmField.FieldID()]
	if !ok {
		return FieldReport{Field: utmField}
	}
	return *result
}

type FieldReport struct {
	Field        data.UTMField
	LACs         map[uint32]*LACReport
	Measurements map[string]*Measurement
}

func NewFieldReport(field data.UTMField) *FieldReport {
	return &FieldReport{
		Field:        field,
		LACs:         make(map[uint32]*LACReport),
		Measurements: make(map[string]*Measurement),
	}
}

func (f *FieldReport) Area() (minLat float64, minLon float64, maxLat float64, maxLon float64) {
	return f.Field.Area()
}

func (f *FieldReport) Add(dataPoint data.DataPoint) {
	lacStats, ok := f.LACs[dataPoint.LAC]
	if !ok {
		lacStats = &LACReport{LAC: dataPoint.LAC}
		f.LACs[dataPoint.LAC] = lacStats
	}
	lacStats.Add(dataPoint)

	measurementKey := dataPoint.MeasurementID()
	measurements, ok := f.Measurements[measurementKey]
	if !ok {
		measurements = &Measurement{ID: measurementKey}
		f.Measurements[measurementKey] = measurements
	}
	measurements.Add(dataPoint)
}

func (f *FieldReport) LACReportsByLAC() []LACReport {
	return f.lacReportsSortedBy(func(lacReports []LACReport) []LACReport {
		// best server first
		slices.SortFunc(lacReports, func(i, j LACReport) int {
			return int(i.LAC) - int(j.LAC)
		})
		return lacReports
	})
}

func (f *FieldReport) LACReportsByRSSI() []LACReport {
	return f.lacReportsSortedBy(func(lacReports []LACReport) []LACReport {
		// best server first, in doubt by LAC
		slices.SortFunc(lacReports, func(i, j LACReport) int {
			return data.SortRSSI(j.AverageRSSI(), i.AverageRSSI(), int(i.LAC)-int(j.LAC))
		})
		return lacReports
	})
}

func (f *FieldReport) lacReportsSortedBy(sort func([]LACReport) []LACReport) []LACReport {
	result := make([]LACReport, 0, len(f.LACs))
	for _, lacReport := range f.LACs {
		result = append(result, *lacReport)
	}
	result = sort(result)
	return result
}

func (f *FieldReport) AverageRSSI() int {
	sum := 0
	count := 0
	for _, measurement := range f.Measurements {
		bestRSSI := measurement.BestRSSI()
		if bestRSSI == data.NoSignal {
			continue
		}
		sum += bestRSSI
		count++
	}
	if count == 0 {
		return data.NoSignal
	}
	return sum / count
}

func (f *FieldReport) AverageGAN() int {
	avgRSSI := f.AverageRSSI()
	if avgRSSI == data.NoSignal {
		return data.NoGAN
	}
	return data.RSSIToGAN(avgRSSI)
}

func (f *FieldReport) AverageSignalLevelDifference() int {
	sum := 0
	count := 0
	for _, measurement := range f.Measurements {
		diff := measurement.SignalLevelDifference()
		if diff != 0 {
			sum += diff
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return sum / count
}

type LACReport struct {
	LAC     uint32
	MinRSSI int
	MaxRSSI int

	rssi []int
}

func (s *LACReport) Add(dataPoint data.DataPoint) {
	if dataPoint.LAC != s.LAC {
		return
	}

	if dataPoint.RSSI == data.NoSignal {
		return
	}

	s.rssi = append(s.rssi, dataPoint.RSSI)

	if len(s.rssi) == 1 {
		s.MinRSSI = dataPoint.RSSI
		s.MaxRSSI = dataPoint.RSSI
	} else {
		if dataPoint.RSSI < s.MinRSSI {
			s.MinRSSI = dataPoint.RSSI
		}
		if dataPoint.RSSI > s.MaxRSSI {
			s.MaxRSSI = dataPoint.RSSI
		}
	}
}

func (c *LACReport) CurrentRSSI() int {
	if len(c.rssi) == 0 {
		return data.NoSignal
	}
	return c.rssi[len(c.rssi)-1]
}

func (s *LACReport) CurrentGAN() int {
	currentRSSI := s.CurrentRSSI()
	if currentRSSI == data.NoSignal {
		return data.NoGAN
	}
	return data.RSSIToGAN(currentRSSI)
}

func (s *LACReport) AverageRSSI() int {
	if len(s.rssi) == 0 {
		return data.NoSignal
	}

	sum := 0
	for _, rssi := range s.rssi {
		sum += rssi
	}
	return sum / len(s.rssi)
}

func (s *LACReport) AverageGAN() int {
	avgRSSI := s.AverageRSSI()
	if avgRSSI == data.NoSignal {
		return data.NoGAN
	}
	return data.RSSIToGAN(avgRSSI)
}

type Measurement struct {
	ID         string
	dataPoints []data.DataPoint
}

func (m *Measurement) Add(dataPoints ...data.DataPoint) {
	if len(dataPoints) > 0 && m.ID == "" && len(m.dataPoints) == 0 {
		m.ID = dataPoints[0].MeasurementID()
	}
	for _, dataPoint := range dataPoints {
		if dataPoint.MeasurementID() != m.ID {
			return
		}
		m.dataPoints = append(m.dataPoints, dataPoint)
		m.dataPoints = data.SortByRSSI(m.dataPoints)
	}
}

func (m *Measurement) BestServer() data.DataPoint {
	if len(m.dataPoints) == 0 {
		return data.ZeroDataPoint
	}
	return m.dataPoints[0]
}

func (m *Measurement) SecondServer() data.DataPoint {
	if len(m.dataPoints) < 2 {
		return data.ZeroDataPoint
	}
	return m.dataPoints[1]
}

func (m *Measurement) BestRSSI() int {
	if len(m.dataPoints) == 0 {
		return data.NoSignal
	}
	return m.dataPoints[0].RSSI
}

func (m *Measurement) SignalLevelDifference() int {
	if len(m.dataPoints) < 2 {
		return 0
	}
	if m.dataPoints[0].RSSI == data.NoSignal || m.dataPoints[1].RSSI == data.NoSignal {
		return 0
	}
	return m.dataPoints[0].RSSI - m.dataPoints[1].RSSI
}
