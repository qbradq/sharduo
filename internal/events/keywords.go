package events

import (
	"strings"

	"github.com/qbradq/sharduo/internal/game"
)

func init() {
	reg("KeywordsBanker", KeywordsBanker)
	reg("KeywordsVendor", KeywordsVendor)
}

// KeywordsBanker handles banker speech triggers.
func KeywordsBanker(receiver, source game.Object, v any) {
	if strings.Contains(strings.ToLower(v.(string)), "bank") {
		OpenBankBox(receiver, source, v)
	}
}

// KeywordsVendor handles common vendor speech triggers.
func KeywordsVendor(receiver, source game.Object, v any) {
	s := strings.ToLower(v.(string))
	if !strings.Contains(s, "vendor") {
		return
	}
	if strings.Contains(s, "buy") {
		VendorBuy(receiver, source, v)
	} else if strings.Contains(s, "sell") {
		VendorSell(receiver, source, v)
	}
}