package util

import (
	"crypto/sha256"
	"encoding/hex"
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

func ParseLocation(strval string) (uo.Point, error) {
	parts := strings.Split(strval, ",")
	if len(parts) != 3 {
		return uo.Point{}, fmt.Errorf("GetLocation(%s) did not find three values", strval)
	}
	var l uo.Point
	v, err := strconv.ParseInt(parts[0], 0, 16)
	if err != nil {
		return uo.Point{}, err
	}
	l.X = int(v)
	v, err = strconv.ParseInt(parts[1], 0, 16)
	if err != nil {
		return uo.Point{}, err
	}
	l.Y = int(v)
	v, err = strconv.ParseInt(parts[2], 0, 8)
	if err != nil {
		return uo.Point{}, err
	}
	l.Z = int(v)
	return l, nil
}

// Hashes a password suitable for the accounts database.
func HashPassword(password string) string {
	hd := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hd[:])
}

// RandomLocation returns a random point within the bounds.
func RandomLocation(b uo.Bounds) uo.Point {
	return uo.Point{
		X: Random(b.X, b.East()),
		Y: Random(b.Y, b.South()),
	}
}
