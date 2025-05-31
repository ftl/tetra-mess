package quality

import "github.com/ftl/tetra-mess/pkg/data"

type QualityReport struct {
	fieldsByUTM map[string]*FieldReport
}

func NewQualityReport() *QualityReport {
	return &QualityReport{
		fieldsByUTM: make(map[string]*FieldReport),
	}
}

func (a *QualityReport) Add(dataPoint data.DataPoint) {
	fieldKey := dataPoint.UTMField().FieldID()

	field, ok := a.fieldsByUTM[fieldKey]
	if !ok {
		field = NewFieldReport(dataPoint.UTMField())
		a.fieldsByUTM[fieldKey] = field
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

	measurementKey := dataPoint.TimeAndSpace()
	measurements, ok := f.Measurements[measurementKey]
	if !ok {
		measurements = &Measurement{ID: measurementKey}
		f.Measurements[measurementKey] = measurements
	}
	measurements.Add(dataPoint)
}

func (f *FieldReport) AverageRSSI() int {
	sum := 0
	count := 0
	for _, measurement := range f.Measurements {
		bestRSSI := measurement.BestRSSI()
		if bestRSSI != data.NoSignal {
			sum += bestRSSI
			count++
		}
	}
	if count == 0 {
		return data.NoSignal
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

type Measurement struct {
	ID         string
	dataPoints []data.DataPoint
}

func (m *Measurement) Add(dataPoint data.DataPoint) {
	if dataPoint.TimeAndSpace() != m.ID {
		return
	}
	m.dataPoints = append(m.dataPoints, dataPoint)
}

func (m *Measurement) BestRSSI() int {
	result := -1000
	for _, dataPoint := range m.dataPoints {
		if dataPoint.RSSI != data.NoSignal && dataPoint.RSSI > result {
			result = dataPoint.RSSI
		}
	}
	if result == -1000 {
		return data.NoSignal
	}
	return result
}
