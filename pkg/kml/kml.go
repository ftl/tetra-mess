package kml

import (
	"fmt"
	"io"

	"github.com/twpayne/go-kml/v3"

	"github.com/ftl/tetra-mess/pkg/data"
	"github.com/ftl/tetra-mess/pkg/quality"
)

func WriteDataPointsAsKML(out io.Writer, name string, dataPoints []data.DataPoint) error {
	elements := make([]kml.Element, 0, len(dataPoints)+1)
	elements = append(elements,
		kml.Name(name),
	)
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
	gan := data.RSSIToGAN(dataPoint.RSSI)
	color := data.GANToColor(gan)
	return kml.Placemark(
		kml.Name(fmt.Sprintf("%d/%x %ddBm", dataPoint.LAC, dataPoint.LAC, dataPoint.RSSI)),
		kml.Description(fmt.Sprintf("LAC: %d<br/>Carrier: %x<br/>RSSI: %ddBm<br/>Cx: %d<br/>GAN: %d", dataPoint.LAC, dataPoint.Carrier, dataPoint.RSSI, dataPoint.Cx, gan)),
		kml.TimeStamp(kml.When(dataPoint.Timestamp)),
		kml.Point(
			kml.Coordinates(kml.Coordinate{Lat: dataPoint.Latitude, Lon: dataPoint.Longitude}),
		),
		kml.Style(
			kml.IconStyle(
				// https://kml4earth.appspot.com/icons.html
				kml.Icon(kml.Href("http://maps.google.com/mapfiles/kml/shapes/placemark_circle.png")),
				kml.Color(color),
			),
			kml.LabelStyle(
				kml.Color(color),
			),
		),
	)
}

func WriteFieldReportsAsKML(out io.Writer, name string, fieldReports []quality.FieldReport) error {
	elements := make([]kml.Element, 0, len(fieldReports)+9)
	elements = append(elements,
		kml.Name(name),
	)
	for gan := data.NoGAN; gan <= 4; gan++ {
		styleID := fmt.Sprintf("gan%d-style", gan)
		style := kml.Style(
			kml.PolyStyle(
				kml.Color(data.GANToColor(gan)),
				kml.Fill(true),
			),
		).WithID(styleID)
		elements = append(elements, style)
	}
	elements = append(elements, fieldReportsToKMLPlacemarks(fieldReports)...)

	doc := kml.KML(
		kml.Document(elements...),
	)

	return doc.WriteIndent(out, "", "  ")
}

func fieldReportsToKMLPlacemarks(fieldReports []quality.FieldReport) []kml.Element {
	result := make([]kml.Element, 0, len(fieldReports))
	for _, fieldStat := range fieldReports {
		minLat, minLon, maxLat, maxLon := fieldStat.Area()
		if minLat == 0 && minLon == 0 && maxLat == 0 && maxLon == 0 {
			continue // Skip fields without valid area
		}
		avgGAN := data.RSSIToGAN(fieldStat.AverageRSSI())
		styleURL := fmt.Sprintf("#gan%d-style", avgGAN)
		placemark := kml.Placemark(
			kml.Name(fmt.Sprintf("Field %s", fieldStat.Field.FieldID())),
			kml.Description(fieldReportDescription(fieldStat)),
			kml.StyleURL(styleURL),
			kml.Polygon(
				kml.OuterBoundaryIs(
					kml.LinearRing(
						kml.Coordinates(
							kml.Coordinate{Lat: maxLat, Lon: minLon},
							kml.Coordinate{Lat: minLat, Lon: minLon},
							kml.Coordinate{Lat: minLat, Lon: maxLon},
							kml.Coordinate{Lat: maxLat, Lon: maxLon},
							kml.Coordinate{Lat: maxLat, Lon: minLon},
						),
					),
				),
			),
		)

		result = append(result, placemark)
	}
	return result
}

func fieldReportDescription(fieldReports quality.FieldReport) string {
	var result string
	result += fmt.Sprintf(`<table>
<tr><th>UTM Field</th><td>%s</td></tr>
<tr><th>Avg RSSI</th><td>%ddBm</td></tr>
<tr><th>Avg GAN</th><td>%d</td></tr>
<tr><th>Avg SLD</th><td>%ddB</td></tr>
</table><br/>`,
		fieldReports.Field.FieldID(),
		fieldReports.AverageRSSI(),
		fieldReports.AverageGAN(),
		fieldReports.AverageSignalLevelDifference(),
	)
	result += fmt.Sprintf(`<table>
<tr>
<th>LAC</th>
<th>Avg RSSI</th>
<th>Avg GAN</th>
<th>Min RSSI</th>
<th>Max RSSI</th>
</tr>`)
	for _, lacStats := range fieldReports.LACReportsByRSSI() {
		result += fmt.Sprintf(`<tr><td>%d</td><td>%d</td><td>%d</td><td>%d</td><td>%d</td></tr>`,
			lacStats.LAC,
			lacStats.AverageRSSI(),
			lacStats.AverageGAN(),
			lacStats.MinRSSI,
			lacStats.MaxRSSI,
		)
	}
	result += `</table><br/>`

	return result
}
