package events

import (
	"strconv"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	reg("RandomHue", randomHue)
	reg("RandomPartialHue", randomPartialHue)
}

func randomHue(r, s, v any) bool {
	ln := v.(string)
	hs := game.ListMember(ln)
	iv, err := strconv.ParseInt(hs, 0, 32)
	if err != nil {
		panic(err)
	}
	hue := uo.Hue(iv)
	if m, ok := r.(*game.Mobile); ok {
		m.Hue = hue
	} else {
		r.(*game.Item).Hue = hue
	}
	return true
}

func randomPartialHue(r, s, v any) bool {
	ln := v.(string)
	hs := game.ListMember(ln)
	iv, err := strconv.ParseInt(hs, 0, 32)
	if err != nil {
		panic(err)
	}
	hue := uo.Hue(iv).SetPartial()
	if m, ok := r.(*game.Mobile); ok {
		m.Hue = hue
	} else {
		r.(*game.Item).Hue = hue
	}
	return true
}
