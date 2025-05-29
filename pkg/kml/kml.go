package kml

import (
	"fmt"
	"io"

	"github.com/twpayne/go-kml/v3"

	"github.com/ftl/tetra-mess/pkg/data"
)

func WriteDataPointsAsKML(out io.Writer, name string, dataPoints []data.DataPoint) error {
	elements := make([]kml.Element, 0, len(dataPoints)+1)
	elements = append(elements, kml.Name(name))
	elements = append(elements, dataPointsToKMLPlacemarks(dataPoints)...)

	doc := kml.KML(
		kml.Document(elements...),
	)

	return doc.WriteIndent(out, "", "  ")
}

func dataPointsToKMLPlacemarks(dataPoints []data.DataPoint) []kml.Element {
	result := make([]kml.Element, 0, len(dataPoints))
	for _, dataPoint := range dataPoints {
		if dataPoint.Latitude == 0 && dataPoint.Longitude == 0 {
			continue // Skip points without valid coordinates
		}

		point := dataPointToKMLPlacemark(dataPoint)
		result = append(result, point)
	}
	return result
}

func dataPointToKMLPlacemark(dataPoint data.DataPoint) kml.Element {
	return kml.Placemark(
		kml.Name(fmt.Sprintf("%d/%x %ddBm", dataPoint.LAC, dataPoint.LAC, dataPoint.RSSI)),
		kml.Description(fmt.Sprintf("LAC: %d\nID: %x\nRSSI: %ddBm\nCSNR: %d", dataPoint.LAC, dataPoint.ID, dataPoint.RSSI, dataPoint.CSNR)),
		kml.TimeStamp(kml.When(dataPoint.Timestamp)),
		kml.Point(
			kml.Coordinates(kml.Coordinate{Lat: dataPoint.Latitude, Lon: dataPoint.Longitude}),
		),
	)
}
