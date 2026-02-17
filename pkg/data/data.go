package data

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/golang/geo/s2"
	"github.com/tzneal/coordconv"
)

const NoSignal = 99
const NoGAN = -3

var NoPosition = Position{}
var ZeroDataPoint = DataPoint{}

type Position struct {
	Latitude   float64
	Longitude  float64
	Satellites int
	Timestamp  time.Time
}

func (p Position) ToUTMField() UTMField {
	return NewUTMField(p.Latitude, p.Longitude)
}

type CellInfo struct {
	LAC     uint32
	Carrier uint32
	RSSI    int
	Cx      int
}

type DataPoint struct {
	Latitude   float64   `json:"lat"`
	Longitude  float64   `json:"lon"`
	Satellites int       `json:"sats"`
	Timestamp  time.Time `json:"ts"`
	LAC        uint32    `json:"lac"`
	Carrier    uint32    `json:"carrier"`
	RSSI       int       `json:"rssi"`
	Cx         int       `json:"cx"`
}

func (dp DataPoint) IsZero() bool {
	return dp == ZeroDataPoint
}

func (dp DataPoint) IsValid() bool {
	return dp.Satellites > 0 && dp.RSSI != NoSignal
}

func (dp DataPoint) IsUsable() bool {
	return IsUsableRSSI(dp.RSSI)
}

func (dp DataPoint) MeasurementID() string {
	data := fmt.Sprintf("%f-%f-%s", dp.Latitude, dp.Longitude, dp.Timestamp.Format(time.RFC3339))
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (dp DataPoint) UTMField() UTMField {
	return NewUTMField(dp.Latitude, dp.Longitude)
}

type UTMField struct {
	Zone   int
	Letter string
	Square string
	East   int
	North  int
}

func NewUTMField(lat float64, lon float64) UTMField {
	mgrs, err := coordconv.DefaultMGRSConverter.ConvertFromGeodetic(s2.LatLngFromDegrees(lat, lon), 5)
	if err != nil {
		panic(fmt.Sprintf("Error converting lat/lon to MGRS: %v", err))
	}

	zone, _ := strconv.Atoi(mgrs[0:2])
	letter := string(mgrs[2])
	square := string(mgrs[3:5])
	east, _ := strconv.Atoi(string(mgrs[5:10]))
	north, _ := strconv.Atoi(string(mgrs[10:]))

	return UTMField{
		Zone:   zone,
		Letter: letter,
		Square: square,
		East:   east,
		North:  north,
	}
}

func (f UTMField) String() string {
	return fmt.Sprintf("%d%s %s %05d %05d", f.Zone, f.Letter, f.Square, f.East, f.North)
}

func (f UTMField) FieldID() string {
	east100 := f.East / 10
	north100 := f.North / 10
	return fmt.Sprintf("%d%s %s %04d %04d", f.Zone, f.Letter, f.Square, east100, north100)
}

func (f UTMField) Area() (minLat float64, minLon float64, maxLat float64, maxLon float64) {
	minEast := (f.East / 100) * 100
	minNorth := (f.North / 100) * 100
	minLatLon, err := coordconv.DefaultMGRSConverter.ConvertToGeodetic(fmt.Sprintf("%d%s%s%05d%05d", f.Zone, f.Letter, f.Square, minEast, minNorth))
	if err != nil {
		panic(fmt.Sprintf("Error converting UTM to min lat/lon: %v", err))
	}

	maxEast := ((f.East / 10) * 10) + 10
	maxNorth := ((f.North / 10) * 10) + 10
	maxLatLon, err := coordconv.DefaultMGRSConverter.ConvertToGeodetic(fmt.Sprintf("%d%s%s%05d%05d", f.Zone, f.Letter, f.Square, maxEast, maxNorth))
	if err != nil {
		panic(fmt.Sprintf("Error converting UTM to min lat/lon: %v", err))
	}

	return float64(minLatLon.Lat), float64(minLatLon.Lng), float64(maxLatLon.Lat), float64(maxLatLon.Lng)
}

func RSSIToGAN(rssi int) int {
	switch {
	case rssi == NoSignal:
		return NoGAN
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

func IsUsableRSSI(rssi int) bool {
	return rssi >= -94
}
