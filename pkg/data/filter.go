package data

import "slices"

type Filter interface {
	Filter(dataPoints []DataPoint) []DataPoint
}

type FilterFunc func(dataPoints []DataPoint) []DataPoint

func (f FilterFunc) Filter(dataPoints []DataPoint) []DataPoint {
	return f(dataPoints)
}

func FilterByLAC(lac uint32) Filter {
	return FilterFunc(func(dataPoints []DataPoint) []DataPoint {
		result := make([]DataPoint, 0, len(dataPoints))
		for _, dp := range dataPoints {
			if dp.LAC == lac {
				result = append(result, dp)
			}
		}
		return result
	})
}

func FilterByCarrier(carrier uint32) Filter {
	return FilterFunc(func(dataPoints []DataPoint) []DataPoint {
		result := make([]DataPoint, 0, len(dataPoints))
		for _, dp := range dataPoints {
			if dp.Carrier == carrier {
				result = append(result, dp)
			}
		}
		return result
	})
}

func FilterBestServer() Filter {
	return FilterFunc(func(dataPoints []DataPoint) []DataPoint {
		byTimeAndSpace := make(map[string][]DataPoint)
		for _, dataPoint := range dataPoints {
			key := dataPoint.MeasurementID()
			byTimeAndSpace[key] = append(byTimeAndSpace[key], dataPoint)
		}

		result := make([]DataPoint, 0, len(byTimeAndSpace))
		for _, dataPointsAtTimeAndSpace := range byTimeAndSpace {
			bestServer := bestServerAtTimeAndSpace(dataPointsAtTimeAndSpace)
			if !bestServer.IsZero() {
				result = append(result, bestServer)
			}
		}
		return SortByTimestamp(result)
	})
}

func bestServerAtTimeAndSpace(dataPoints []DataPoint) DataPoint {
	if len(dataPoints) == 0 {
		return ZeroDataPoint
	}

	result := dataPoints[0]
	for _, dataPoint := range dataPoints {
		if dataPoint.RSSI > result.RSSI || (dataPoint.RSSI == result.RSSI && dataPoint.Cx > result.Cx) {
			result = dataPoint
		}
	}
	if !result.IsValid() {
		return ZeroDataPoint
	}
	return result
}

func SortByTimestamp(dataPoints []DataPoint) []DataPoint {
	slices.SortFunc(dataPoints, func(i, j DataPoint) int {
		return int(i.Timestamp.Sub(j.Timestamp).Seconds())
	})
	return dataPoints
}

func SortByRSSI(dataPoints []DataPoint) []DataPoint {
	// best server first
	slices.SortFunc(dataPoints, func(i, j DataPoint) int {
		return SortRSSI(j.RSSI, i.RSSI, int(i.Timestamp.Sub(j.Timestamp).Seconds()))
	})
	return dataPoints
}

func SortRSSI(rssi1, rssi2, alt int) int {
	if rssi1 == NoSignal && rssi2 == NoSignal {
		return alt // both are no signal
	}
	if rssi1 == NoSignal {
		return -1 // rssi1 is better
	}
	if rssi2 == NoSignal {
		return 1 // rssi2 is better
	}
	return rssi1 - rssi2
}

func MapByUTMField(dataPoints []DataPoint) map[string][]DataPoint {
	result := make(map[string][]DataPoint)
	for _, dataPoint := range dataPoints {
		field := dataPoint.UTMField()
		key := field.FieldID()
		result[key] = append(result[key], dataPoint)
	}
	return result
}

func MapByMeasurement(dataPoints []DataPoint) map[string][]DataPoint {
	result := make(map[string][]DataPoint)
	for _, dataPoint := range dataPoints {
		key := dataPoint.MeasurementID()
		result[key] = append(result[key], dataPoint)
	}
	return result
}
