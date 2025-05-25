package scanner

import (
	"context"
	"time"

	"github.com/ftl/tetra-pei/com"
	"github.com/ftl/tetra-pei/ctrl"

	"github.com/ftl/tetra-mess/pkg/data"
)

func ScanSignalAndPosition(ctx context.Context, radio *com.COM) ([]data.DataPoint, error) {
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
		return []data.DataPoint{{
			Latitude:   lat,
			Longitude:  lon,
			Satellites: sats,
			Timestamp:  timestamp,
			RSSI:       dbm,
		}}, nil
	}

	result := make([]data.DataPoint, 0, len(cellInfos))
	for _, cellInfo := range cellInfos {
		dataPoint := data.DataPoint{
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
