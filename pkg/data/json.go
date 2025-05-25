package data

import (
	"encoding/json"
	"fmt"
	"strings"
)

func DataPointToJSON(dataPoint DataPoint) string {
	encoded, _ := json.Marshal(dataPoint)
	return string(encoded)
}

func IsJSONLine(line string) bool {
	return strings.HasPrefix(line, "{") && strings.HasSuffix(line, "}") && (strings.Contains(line, "\"") || !strings.Contains(line, "'")) && strings.Contains(line, ":")
}

func ParseJSONLine(line string) (DataPoint, error) {
	reader := json.NewDecoder(strings.NewReader(line))
	var dataPoint DataPoint
	err := reader.Decode(&dataPoint)
	if err != nil {
		return DataPoint{}, fmt.Errorf("error parsing JSON line: %w", err)
	}

	return dataPoint, nil
}
