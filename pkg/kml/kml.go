package kml

import (
	"fmt"
	"image/color"
	"io"

	"github.com/twpayne/go-kml/v3"

	"github.com/ftl/tetra-mess/pkg/data"
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
