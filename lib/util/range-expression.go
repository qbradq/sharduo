package util

import (
	"log"
	"strconv"
	"strings"
)

// RangeExpression parses a string expression and returns a number, possibly
// randomly selected from a set or range.
func RangeExpression(e string) int {
	f := func(s string) int {
		if len(s) < 1 {
			return 0
		}
		i, err := strconv.ParseInt(s, 0, 32)
		if err != nil {
			log.Printf("error: parsing range expression \"%s\": %s", e, err.Error())
			i = 0
		}
		return int(i)
	}
	var ret int
	var set []int
	for _, es := range strings.Split(e, ",") {
		parts := strings.Split(es, "-")
		if len(parts) == 1 {
			set = append(set, f(parts[0]))
		} else if len(parts) == 2 {
			p1 := f(parts[0])
			p2 := f(parts[1])
			for i := p1; i <= p2; i++ {
				set = append(set, i)
			}
		} else {
			log.Printf("error: parsing range expression \"%s\"", e)
		}
	}
	if len(set) > 0 {
		ret = set[Random(0, len(set)-1)]
	}
	return ret
}
