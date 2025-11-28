package scanner

import (
	"context"
	"time"

	"github.com/ftl/tetra-cli/pkg/radio"
	"github.com/ftl/tetra-pei/ctrl"

	"github.com/ftl/tetra-mess/pkg/data"
	"github.com/ftl/tetra-mess/pkg/quality"
)

type Logger func(string, ...any)

type DataPoint struct {
	Position    data.Position
	Measurement quality.Measurement
}

func ScanSignalAndPosition(ctx context.Context, pei radio.PEI, log Logger) (data.Position, []data.DataPoint) {
	lat, lon, sats, timestamp, err := ctrl.RequestGPSPosition(ctx, pei)
	if err != nil {
		log("cannot read GPS position: %v", err)
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
		log("cannot read signal strength: %v", err)
		dbm = 0
	}

	cellInfos, err := RequestCellListInformation(ctx, pei)
	if err != nil {
		log("cannot read cell list information: %v", err)
		return position,
			[]data.DataPoint{{
				Latitude:   lat,
				Longitude:  lon,
				Satellites: sats,
				Timestamp:  timestamp,
				RSSI:       dbm,
			}}
	}

	dataPoints := make([]data.DataPoint, 0, len(cellInfos))
	for _, cellInfo := range cellInfos {
		dataPoint := data.DataPoint{
			Latitude:   lat,
			Longitude:  lon,
			Satellites: sats,
			Timestamp:  timestamp,
			LAC:        cellInfo.LAC,
			Carrier:    cellInfo.Carrier,
			RSSI:       cellInfo.RSSI,
			Cx:         cellInfo.Cx,
		}
		dataPoints = append(dataPoints, dataPoint)
	}
	return position, dataPoints
}
