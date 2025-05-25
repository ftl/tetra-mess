package data

import (
	"strconv"
	"strings"
)

func ParseDecOrHex(s string) (uint32, error) {
	var result uint64
	var err error
	if strings.ContainsAny(s, "ABCDEFabcdef") {
		result, err = strconv.ParseUint(s, 16, 32)
	} else {
		result, err = strconv.ParseUint(s, 10, 32)
	}
	return uint32(result), err
}
