package scanner

import (
	"context"
	"time"

	"github.com/ftl/tetra-pei/com"
	"github.com/ftl/tetra-pei/ctrl"
)

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

func ScanSignalAndPosition(ctx context.Context, radio *com.COM) ([]DataPoint, error) {
	lat, lon, sats, timestamp, err := ctrl.RequestGPSPosition(ctx, radio)
	if err != nil {
		lat = 0
		lon = 0
		sats = 0
		timestamp = time.Now().UTC()
	}

	dbm, err := ctrl.RequestSignalStrength(ctx, radio)
	if err != nil {
		dbm = 0
	}

	cellInfos, err := RequestCellListInformation(ctx, radio)
	if err != nil {
		return []DataPoint{{
			Latitude:   lat,
			Longitude:  lon,
			Satellites: sats,
			Timestamp:  timestamp,
			RSSI:       dbm,
		}}, nil
	}

	result := make([]DataPoint, 0, len(cellInfos))
	for _, cellInfo := range cellInfos {
		dataPoint := DataPoint{
			Latitude:   lat,
			Longitude:  lon,
			Satellites: sats,
			Timestamp:  timestamp,
			LAC:        cellInfo.LAC,
			ID:         cellInfo.ID,
			RSSI:       cellInfo.RSSI,
			CSNR:       cellInfo.CSNR,
		}
		result = append(result, dataPoint)
	}
	return result, nil
}
