package events

import (
	"log"
	"strings"

	"github.com/qbradq/sharduo/internal/game"
)

func init() {
	reg("KeywordsBanker", keywordsBanker)
	reg("KeywordsCommand", keywordsCommand)
	reg("KeywordsStablemaster", keywordsStableMaster)
	reg("KeywordsVendor", keywordsVendor)
}

// speechTarget returns true if the given speech event refers to this object.
// The second return value is the list of words following the name of this
// object.
func speechTarget(names []string, r, s, v any) (bool, []string) {
	rm := r.(*game.Mobile)
	sm := s.(*game.Mobile)
	names = append(names, strings.ToLower(rm.Name))
	if rm.ControlMaster == sm {
		names = append(names, "all")
	}
	line := strings.ToLower(v.(string))
	words := strings.Split(line, " ")
	for i, word := range words {
		for _, name := range names {
			if word == name {
				if i >= len(words)-1 {
					return true, nil
				}
				return true, words[i+1:]
			}
		}
	}
	return false, words
}

// doKeywords handles the available keywords in a standardized way
func doKeywords(hotWords []string, receiver, source any, words []string) bool {
	for i, w := range words {
		for _, hw := range hotWords {
			if w == hw {
				fn, found := keywordEvents[w]
				if found {
					if i >= len(words)-1 {
						return fn(receiver, source, "")
					} else {
						return fn(receiver, source, words[i+1])
					}
				} else {
					log.Printf("error: hot keyword given without handler \"%s\"", w)
					return false
				}
			}
		}
	}
	return false
}

// keywordsBanker handles banker speech triggers.
func keywordsBanker(receiver, source, v any) bool {
	words := strings.Split(v.(string), " ")
	return doKeywords([]string{
		"balance",
		"bank",
		"check",
		"deposit",
		"withdraw",
	}, receiver, source, words)
}

// keywordsVendor handles common vendor speech triggers.
func keywordsVendor(receiver, source, v any) bool {
	f, words := speechTarget([]string{"vendor"}, receiver, source, v)
	if !f {
		return false
	}
	return doKeywords([]string{
		"buy",
		"sell",
	}, receiver, source, words)
}

// keywordsStableMaster handles stablemaster speech triggers.
func keywordsStableMaster(receiver, source, v any) bool {
	f, words := speechTarget([]string{"vendor"}, receiver, source, v)
	if !f {
		return doKeywords([]string{
			"stable",
			"claim",
		}, receiver, source, words)
	}
	return doKeywords([]string{
		"buy",
		"stable",
		"claim",
	}, receiver, source, words)
}

// keywordsCommand handles command-able creature speech triggers.
func keywordsCommand(receiver, source, v any) bool {
	f, words := speechTarget(nil, receiver, source, v)
	if !f {
		return false
	}
	return doKeywords([]string{
		"come",
		"drop",
		"follow",
		"release",
		"stay",
		"stop",
	}, receiver, source, words)
}

// keywordEvents maps keywords to the event handlers they belong to
var keywordEvents = map[string]eventHandler{
	"balance":  bankBalance,
	"bank":     openBankBox,
	"buy":      vendorBuy,
	"check":    bankCheck,
	"claim":    claimAllPets,
	"come":     commandFollowMe,
	"deposit":  bankDeposit,
	"drop":     commandDrop,
	"follow":   commandFollow,
	"release":  commandRelease,
	"sell":     vendorSell,
	"stable":   stablePet,
	"stay":     commandStay,
	"stop":     commandStay,
	"withdraw": bankWithdraw,
}
