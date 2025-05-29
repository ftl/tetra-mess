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

func FilterByID(id uint32) Filter {
	return FilterFunc(func(dataPoints []DataPoint) []DataPoint {
		result := make([]DataPoint, 0, len(dataPoints))
		for _, dp := range dataPoints {
			if dp.ID == id {
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
			key := dataPoint.TimeAndSpace()
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
		if dataPoint.RSSI > result.RSSI || (dataPoint.RSSI == result.RSSI && dataPoint.CSNR > result.CSNR) {
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
