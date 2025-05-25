package data

import "time"

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
