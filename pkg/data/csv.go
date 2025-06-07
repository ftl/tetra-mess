package data

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func DataPointToCSV(dataPoint DataPoint) string {
	return fmt.Sprintf("%s,%f,%f,%d,%d,%x,%d,%d",
		dataPoint.Timestamp.Format(time.RFC3339),
		dataPoint.Latitude,
		dataPoint.Longitude,
		dataPoint.Satellites,
		dataPoint.LAC,
		dataPoint.Carrier,
		dataPoint.RSSI,
		dataPoint.Cx)
}

func IsCSVLine(line string) bool {
	return strings.Contains(line, ",") && !strings.Contains(line, "{") && !strings.Contains(line, "}") && !strings.Contains(line, "\"") && !strings.Contains(line, "'")
}

func ParseCSVLine(line string) (DataPoint, error) {
	reader := csv.NewReader(strings.NewReader(line))
	fields, err := reader.Read()
	if err != nil {
		return DataPoint{}, fmt.Errorf("error reading CSV line: %w", err)
	}
	if len(fields) != 8 {
		return DataPoint{}, fmt.Errorf("expected 8 fields in CSV line, got %d", len(fields))
	}

	timestamp, err := time.Parse(time.RFC3339, fields[0])
	if err != nil {
		return DataPoint{}, fmt.Errorf("error parsing timestamp: %w", err)
	}
	lat, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return DataPoint{}, fmt.Errorf("error parsing latitude: %w", err)
	}
	lon, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return DataPoint{}, fmt.Errorf("error parsing longitude: %w", err)
	}
	sats, err := strconv.Atoi(fields[3])
	if err != nil {
		return DataPoint{}, fmt.Errorf("error parsing satellites: %w", err)
	}
	lac, err := ParseDecOrHex(fields[4])
	if err != nil {
		return DataPoint{}, fmt.Errorf("error parsing LAC: %w", err)
	}
	carrier, err := ParseDecOrHex(fields[5])
	if err != nil {
		return DataPoint{}, fmt.Errorf("error parsing carrier: %w", err)
	}
	rssi, err := strconv.Atoi(fields[6])
	if err != nil {
		return DataPoint{}, fmt.Errorf("error parsing RSSI: %w", err)
	}
	cx, err := strconv.Atoi(fields[7])
	if err != nil {
		return DataPoint{}, fmt.Errorf("error parsing Cx: %w", err)
	}

	return DataPoint{
		Timestamp:  timestamp,
		Latitude:   lat,
		Longitude:  lon,
		Satellites: sats,
		LAC:        lac,
		Carrier:    carrier,
		RSSI:       rssi,
		Cx:         cx,
	}, nil
}
