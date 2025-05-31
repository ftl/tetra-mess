package kml

import (
	"fmt"
	"image/color"
	"io"

	"github.com/twpayne/go-kml/v3"

	"github.com/ftl/tetra-mess/pkg/data"
	"github.com/ftl/tetra-mess/pkg/quality"
)

var (
	GANMinus2Color = color.RGBA{R: 139, G: 0, B: 0, A: 255}
	GANMinus1Color = color.RGBA{R: 220, G: 20, B: 60, A: 255}
	GAN0Color      = color.RGBA{R: 255, G: 140, B: 0, A: 255}
	GAN1Color      = color.RGBA{R: 255, G: 215, B: 0, A: 255}
	GAN2Color      = color.RGBA{R: 154, G: 205, B: 50, A: 255}
	GAN3Color      = color.RGBA{R: 34, G: 139, B: 34, A: 255}
	GAN4Color      = color.RGBA{R: 0, G: 100, B: 0, A: 255}
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
	color := ganToColor(gan)
	return kml.Placemark(
		kml.Name(fmt.Sprintf("%d/%x %ddBm", dataPoint.LAC, dataPoint.LAC, dataPoint.RSSI)),
		kml.Description(fmt.Sprintf("LAC: %d<br/>ID: %x<br/>RSSI: %ddBm<br/>CSNR: %d<br/>GAN: %d", dataPoint.LAC, dataPoint.ID, dataPoint.RSSI, dataPoint.CSNR, gan)),
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

func WriteFieldStatsAsKML(out io.Writer, name string, fieldReports []quality.FieldReport) error {
	elements := make([]kml.Element, 0, len(fieldReports)+9)
	elements = append(elements,
		kml.Name(name),
	)
	for gan := -3; gan <= 4; gan++ {
		styleID := fmt.Sprintf("gan%d-style", gan)
		style := kml.Style(
			kml.PolyStyle(
				kml.Color(ganToColor(gan)),
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
	result += fmt.Sprintf(`Field: %s<br/>
Average RSSI: %ddBm, Average GAN: %d<br/>`,
		fieldReports.Field.FieldID(),
		fieldReports.AverageRSSI(),
		data.RSSIToGAN(fieldReports.AverageRSSI()))

	for _, lacStats := range fieldReports.LACs {
		result += fmt.Sprintf(`LAC: %d<br/>
Average RSSI: %ddBm, Average GAN: %d<br/>Min RSSI: %ddBm, Max RSSI: %ddBm<br/>`,
			lacStats.LAC,
			lacStats.AverageRSSI(),
			data.RSSIToGAN(lacStats.AverageRSSI()),
			lacStats.MinRSSI,
			lacStats.MaxRSSI,
		)
	}

	return result
}

func ganToColor(gan int) color.Color {
	colors := []color.Color{
		GANMinus2Color,
		GANMinus1Color,
		GAN0Color,
		GAN1Color,
		GAN2Color,
		GAN3Color,
		GAN4Color,
	}
	if gan < -2 || gan > 4 {
		return color.Black
	}
	return colors[gan+2]
}
