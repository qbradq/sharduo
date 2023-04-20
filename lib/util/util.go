package util

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/qbradq/sharduo/lib/uo"
)

func ParseTagLine(line string) (string, string, error) {
	parts := strings.SplitN(line, "=", 2)
	var key, value string
	if len(parts) == 1 {
		key = strings.TrimSpace(parts[0])
		value = ""
	} else if len(parts) == 2 {
		key = strings.TrimSpace(parts[0])
		value = strings.TrimSpace(parts[1])
	} else if len(parts) != 2 {
		return "", "", errors.New("syntax error")
	}
	return key, value, nil
}

func ParseLocation(strval string) (uo.Location, error) {
	parts := strings.Split(strval, ",")
	if len(parts) != 3 {
		return uo.Location{}, fmt.Errorf("GetLocation(%s) did not find three values", strval)
	}
	var l uo.Location
	v, err := strconv.ParseInt(parts[0], 0, 16)
	if err != nil {
		return uo.Location{}, err
	}
	l.X = int16(v)
	v, err = strconv.ParseInt(parts[1], 0, 16)
	if err != nil {
		return uo.Location{}, err
	}
	l.Y = int16(v)
	v, err = strconv.ParseInt(parts[2], 0, 8)
	if err != nil {
		return uo.Location{}, err
	}
	l.Z = int8(v)
	return l, nil
}
