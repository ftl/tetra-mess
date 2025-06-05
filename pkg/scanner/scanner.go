package scanner

import (
	"context"
	"time"

	"github.com/ftl/tetra-pei/ctrl"

	"github.com/ftl/tetra-mess/pkg/data"
	"github.com/ftl/tetra-mess/pkg/radio"
)

func ScanSignalAndPosition(ctx context.Context, pei radio.PEI) (data.Position, []data.DataPoint, error) {
	lat, lon, sats, timestamp, err := ctrl.RequestGPSPosition(ctx, pei)
	if err != nil {
		lat = 0
		lon = 0
		sats = 0
		timestamp = time.Now().UTC()
	}

	position := data.Position{
		Latitude:   lat,
		Longitude:  lon,
		Satellites: sats,
		Timestamp:  timestamp,
	}

	dbm, err := ctrl.RequestSignalStrength(ctx, pei)
	if err != nil {
		dbm = 0
	}

	cellInfos, err := RequestCellListInformation(ctx, pei)
	if err != nil {
		return position,
			[]data.DataPoint{{
				Latitude:   lat,
				Longitude:  lon,
				Satellites: sats,
				Timestamp:  timestamp,
				RSSI:       dbm,
			}}, nil
	}

	dataPoints := make([]data.DataPoint, 0, len(cellInfos))
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
		dataPoints = append(dataPoints, dataPoint)
	}
	return position, dataPoints, nil
}
