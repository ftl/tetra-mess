package data

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"
)

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
	return dp.Satellites > 0 && dp.RSSI != 99
}

func (dp DataPoint) TimeAndSpace() string {
	data := fmt.Sprintf("%f-%f-%s", dp.Latitude, dp.Longitude, dp.Timestamp.Format(time.RFC3339))
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}
