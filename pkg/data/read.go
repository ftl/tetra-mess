package data

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strings"
)

func ReadDataPoints(in io.Reader) ([]DataPoint, error) {
	lines, err := ReadLines(in)
	if err != nil {
		return nil, err
	}

	result := make([]DataPoint, 0, len(lines))
	for i, line := range lines {
		var dataPoint DataPoint
		switch {
		case IsCSVLine(line):
			dataPoint, err = ParseCSVLine(line)
		case IsJSONLine(line):
			dataPoint, err = ParseJSONLine(line)
		default:
			err = fmt.Errorf("unknown line format: %s", line)
		}
		if err != nil {
			log.Printf("Error parsing line %d: %v\n", i+1, err)
			continue
		}
		result = append(result, dataPoint)
	}
	return result, nil
}

func ReadLines(in io.Reader) ([]string, error) {
	result := make([]string, 0)
	lineScanner := bufio.NewScanner(in)
	for lineScanner.Scan() {
		line := lineScanner.Text()
		line = strings.TrimSpace(line)
		if line == "" || line[0] == '#' {
			continue
		}
		result = append(result, line)
	}
	return result, lineScanner.Err()
}
