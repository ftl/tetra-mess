package gpx

import (
	"fmt"
	"io"

	"github.com/tkrajina/gpxgo/gpx"

	"github.com/ftl/tetra-mess/pkg/data"
)

func WriteDataPointsAsGPX(out io.Writer, name string, dataPoints []data.DataPoint) error {
	waypoints := dataPointsToGPXPoints(dataPoints)
	track := dataPointsToGPXTrack(dataPoints)
	result := gpx.GPX{
		Version:   "1.1",
		Creator:   "tetra-mess",
		Name:      name,
		Tracks:    []gpx.GPXTrack{track},
		Waypoints: waypoints,
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
	gpxPoints := dataPointsToGPXPoints(dataPoints)
	segment := gpx.GPXTrackSegment{
		Points: gpxPoints,
	}

	track := gpx.GPXTrack{
		Name: "Converted Track",
	}
	track.Segments = []gpx.GPXTrackSegment{
		segment,
	}

	return track
}

func dataPointsToGPXPoints(dataPoints []data.DataPoint) []gpx.GPXPoint {
	result := make([]gpx.GPXPoint, 0, len(dataPoints))
	for _, dataPoint := range dataPoints {
		if dataPoint.Latitude == 0 && dataPoint.Longitude == 0 {
			continue // Skip points without valid coordinates
		}

		point := dataPointToGPXPoint(dataPoint)
		result = append(result, point)
	}
	return result
}

func dataPointToGPXPoint(dataPoint data.DataPoint) gpx.GPXPoint {
	gan := data.RSSIToGAN(dataPoint.RSSI)
	result := gpx.GPXPoint{
		Point: gpx.Point{
			Latitude:  dataPoint.Latitude,
			Longitude: dataPoint.Longitude,
		},
		Name:        fmt.Sprintf("%d/%x %ddBm", dataPoint.LAC, dataPoint.LAC, dataPoint.RSSI),
		Description: fmt.Sprintf("LAC: %d\nCarrier: %x\nRSSI: %ddBm\nCx: %d\nGAN: %d", dataPoint.LAC, dataPoint.Carrier, dataPoint.RSSI, dataPoint.Cx, gan),
		Timestamp:   dataPoint.Timestamp,
	}
	result.Satellites.SetValue(dataPoint.Satellites)
	return result
}
