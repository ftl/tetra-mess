package gpx

import (
	"fmt"
	"io"

	"github.com/tkrajina/gpxgo/gpx"

	"github.com/ftl/tetra-mess/pkg/data"
)

func WriteAsGPX(out io.Writer, dataPoints []data.DataPoint) error {
	track := dataPointsToGPXTrack(dataPoints)
	result := gpx.GPX{
		Version: "1.1",
		Creator: "tetra-mess",
		Tracks:  []gpx.GPXTrack{track},
	}

	bytes, err := gpx.ToXml(&result, gpx.ToXmlParams{
		Indent:  true,
		Version: "1.1",
	})
	if err != nil {
		return fmt.Errorf("error converting data points to GPX XML: %w", err)
	}

	_, err = out.Write(bytes)
	if err != nil {
		return fmt.Errorf("error writing GPX XML to output: %w", err)
	}

	return nil
}

func dataPointsToGPXTrack(dataPoints []data.DataPoint) gpx.GPXTrack {
	segment := gpx.GPXTrackSegment{
		Points: make([]gpx.GPXPoint, 0, len(dataPoints)),
	}
	for _, dataPoint := range dataPoints {
		if dataPoint.Latitude == 0 && dataPoint.Longitude == 0 {
			continue // Skip points without valid coordinates
		}

		point := dataPointToGPXPoint(dataPoint)
		segment.Points = append(segment.Points, point)
	}

	track := gpx.GPXTrack{
		Name: "Converted Track",
	}
	track.Segments = []gpx.GPXTrackSegment{
		segment,
	}

	return track
}

func dataPointToGPXPoint(dataPoint data.DataPoint) gpx.GPXPoint {
	result := gpx.GPXPoint{
		Point: gpx.Point{
			Latitude:  dataPoint.Latitude,
			Longitude: dataPoint.Longitude,
		},
		Name:        fmt.Sprintf("%ddBm C %d", dataPoint.RSSI, dataPoint.CSNR),
		Description: fmt.Sprintf("LAC: %d, ID: %d", dataPoint.LAC, dataPoint.ID),
		Timestamp:   dataPoint.Timestamp,
	}
	result.Satellites.SetValue(dataPoint.Satellites)
	return result
}
