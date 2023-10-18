package events

import (
	"log"
	"strings"

	"github.com/qbradq/sharduo/internal/game"
)

func init() {
	reg("KeywordsBanker", KeywordsBanker)
	reg("KeywordsStablemaster", KeywordsStablemaster)
	reg("KeywordsVendor", KeywordsVendor)
}

// doKeywords handles the available keywords in a standardized way
func doKeywords(hotWords []string, receiver, source game.Object, v any) {
	line := v.(string)
	words := strings.Split(line, " ")
	for _, w := range words {
		for _, hw := range hotWords {
			if w == hw {
				fn, found := keywordEvents[w]
				if found {
					fn(receiver, source, v)
				} else {
					log.Printf("hot keyword given without handler \"%s\"", w)
				}
				return
			}
		}
	}
}

// KeywordsBanker handles banker speech triggers.
func KeywordsBanker(receiver, source game.Object, v any) bool {
	doKeywords([]string{
		"bank",
	}, receiver, source, v)
	return true
}

// KeywordsVendor handles common vendor speech triggers.
func KeywordsVendor(receiver, source game.Object, v any) bool {
	doKeywords([]string{
		"buy",
		"sell",
	}, receiver, source, v)
	return true
}

// KeywordsStablemaster handles stablemaster speech triggers.
func KeywordsStablemaster(receiver, source game.Object, v any) bool {
	doKeywords([]string{
		"buy",
		"stable",
		"claim",
	}, receiver, source, v)
	return true
}

// keywordEvents maps keywords to the event handlers they belong to
var keywordEvents = map[string]EventHandler{
	"bank":   OpenBankBox,
	"buy":    VendorBuy,
	"sell":   VendorSell,
	"stable": StablePet,
	"claim":  ClaimAllPets,
}
