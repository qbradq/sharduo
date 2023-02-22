package events

import (
	"strings"

	"github.com/qbradq/sharduo/internal/game"
)

func init() {
	reg("KeywordsBanker", KeywordsBanker)
}

// KeywordsBanker handles banker speech triggers.
func KeywordsBanker(receiver, source game.Object, v any) {
	if strings.Contains(strings.ToLower(v.(string)), "bank") {
		OpenBankBox(source, nil, v)
	}
}
