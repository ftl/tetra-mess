package scanner

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/ftl/tetra-cli/pkg/radio"

	"github.com/ftl/tetra-mess/pkg/data"
)

var cellListResponseHeader = regexp.MustCompile(`^\+GCLI: (\d+)$`)

func RequestCellListInformation(ctx context.Context, pei radio.PEI) ([]data.CellInfo, error) {
	response, err := pei.AT(ctx, "AT+GCLI?")
	if err != nil {
		return nil, err
	}
	if len(response) == 0 {
		return nil, fmt.Errorf("empty response received")
	}

	headerParts := cellListResponseHeader.FindStringSubmatch(response[0])
	if len(headerParts) != 2 {
		return nil, fmt.Errorf("invalid response header: %s", response[0])
	}
	count, err := strconv.Atoi(headerParts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid response count: %w", err)
	}
	if len(response) != count+1 {
		return nil, fmt.Errorf("invalid response length: %d != %d", len(response), count+1)
	}

	result := make([]data.CellInfo, 0, count)
	for _, line := range response[1:] {
		cellInfo, err := parseCellInfo(line)
		if err != nil {
			log.Printf("invalid cell info line: %v", err) // TODO: print to stderr
			continue
		}
		result = append(result, cellInfo)
	}

	return result, nil
}

var cellInfoLineExpression = regexp.MustCompile(`^(\d+),([abcdef0123456789]+),(-?\d+),(-?\d+)`)

func parseCellInfo(line string) (data.CellInfo, error) {
	line = strings.ToLower(strings.TrimSpace(line))
	parts := cellInfoLineExpression.FindStringSubmatch(line)
	if len(parts) != 5 {
		return data.CellInfo{}, fmt.Errorf("invalid cell info line: %s", line)
	}

	lac, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return data.CellInfo{}, fmt.Errorf("invalid LAC: %w", err)
	}

	carrier, err := strconv.ParseUint(parts[2], 16, 32)
	if err != nil {
		return data.CellInfo{}, fmt.Errorf("invalid carrier: %w", err)
	}

	rawRSSI, err := strconv.Atoi(parts[3])
	if err != nil {
		return data.CellInfo{}, fmt.Errorf("invalid RSSI: %w", err)
	}
	var rssi int
	if rawRSSI != data.NoSignal {
		rssi = -113 + (rawRSSI * 2)
	} else {
		rssi = data.NoSignal
	}

	cx, err := strconv.Atoi(parts[4])
	if err != nil {
		return data.CellInfo{}, fmt.Errorf("invalid Cx: %w", err)
	}

	return data.CellInfo{
		LAC:     uint32(lac),
		Carrier: uint32(carrier),
		RSSI:    rssi,
		Cx:      cx,
	}, nil
}
