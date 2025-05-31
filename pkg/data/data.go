package data

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/im7mortal/UTM"
)

const NoSignal = 99

var ZeroDataPoint = DataPoint{}

type CellInfo struct {
	LAC  uint32
	ID   uint32
	RSSI int
	CSNR int
}

type DataPoint struct {
	Latitude   float64   `json:"lat"`
	Longitude  float64   `json:"lon"`
	Satellites int       `json:"sats"`
	Timestamp  time.Time `json:"ts"`
	LAC        uint32    `json:"lac"`
	ID         uint32    `json:"id"`
	RSSI       int       `json:"rssi"`
	CSNR       int       `json:"csnr"`
}

func (dp DataPoint) IsZero() bool {
	return dp == ZeroDataPoint
}

func (dp DataPoint) IsValid() bool {
	return dp.Satellites > 0 && dp.RSSI != NoSignal
}

func (dp DataPoint) TimeAndSpace() string {
	data := fmt.Sprintf("%f-%f-%s", dp.Latitude, dp.Longitude, dp.Timestamp.Format(time.RFC3339))
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (dp DataPoint) UTMField() UTMField {
	return NewUTMField(dp.Latitude, dp.Longitude)
}

type UTMField struct {
	East   float64
	North  float64
	Zone   int
	Letter string
}

func NewUTMField(lat float64, lon float64) UTMField {
	east, north, zone, letter, err := UTM.FromLatLon(lat, lon, false)
	if err != nil {
		panic(fmt.Sprintf("Error converting lat/lon to UTM: %v", err))
	}

	return UTMField{
		East:   east,
		North:  north,
		Zone:   zone,
		Letter: letter,
	}
}

func (f UTMField) String() string {
	return fmt.Sprintf("%d%s %06.0f %06.0f", f.Zone, f.Letter, f.East, f.North)
}

func (f UTMField) FieldID() string {
	east100 := fmt.Sprintf("%06.0f", f.East)[:4]
	north100 := fmt.Sprintf("%07.0f", f.North)[:5]
	return fmt.Sprintf("%d%s %s %s", f.Zone, f.Letter, east100, north100)
}

func (f UTMField) Area() (minLat float64, minLon float64, maxLat float64, maxLon float64) {
	minEast := float64(int64(f.East/100) * 100)
	minNorth := float64(int64(f.North/100) * 100)
	minLat, minLon, err := UTM.ToLatLon(minEast, minNorth, f.Zone, f.Letter)
	if err != nil {
		panic(fmt.Sprintf("Error converting UTM to min lat/lon: %v", err))
	}

	maxEast := float64(int64(f.East/100)*100) + 100
	maxNorth := float64(int64(f.North/100)*100) + 100
	maxLat, maxLon, err = UTM.ToLatLon(maxEast, maxNorth, f.Zone, f.Letter)
	if err != nil {
		panic(fmt.Sprintf("Error converting UTM to max lat/lon: %v", err))
	}
	return
}

func RSSIToGAN(rssi int) int {
	switch {
	case rssi == NoSignal:
		return -3
	case rssi < -103:
		return -2
	case rssi < -97:
		return -1
	case rssi < -94:
		return 0
	case rssi < -88:
		return 1
	case rssi < -85:
		return 2
	case rssi < -79:
		return 3
	default:
		return 4
	}
}
