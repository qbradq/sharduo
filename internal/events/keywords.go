package events

import (
	"log"
	"strings"

	"github.com/qbradq/sharduo/internal/game"
)

func init() {
	reg("KeywordsBanker", KeywordsBanker)
	reg("KeywordsCommand", KeywordsCommand)
	reg("KeywordsStablemaster", KeywordsStablemaster)
	reg("KeywordsVendor", KeywordsVendor)
}

// speechTarget returns true if the given speech event refers to this object.
// The second return value is the list of words following the name of this
// object.
func speechTarget(names []string, receiver, source game.Object, v any) (bool, []string) {
	names = append(names, strings.ToLower(receiver.Name()))
	if rm, ok := receiver.(game.Mobile); ok && rm.ControlMaster() != nil &&
		rm.ControlMaster().Serial() == source.Serial() {
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
func doKeywords(hotWords []string, receiver, source game.Object, words []string) bool {
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
					log.Printf("hot keyword given without handler \"%s\"", w)
					return false
				}
			}
		}
	}
	return false
}

// KeywordsBanker handles banker speech triggers.
func KeywordsBanker(receiver, source game.Object, v any) bool {
	words := strings.Split(v.(string), " ")
	return doKeywords([]string{
		"bank",
	}, receiver, source, words)
}

// KeywordsVendor handles common vendor speech triggers.
func KeywordsVendor(receiver, source game.Object, v any) bool {
	f, words := speechTarget([]string{"vendor"}, receiver, source, v)
	if !f {
		return false
	}
	return doKeywords([]string{
		"buy",
		"sell",
	}, receiver, source, words)
}

// KeywordsStablemaster handles stablemaster speech triggers.
func KeywordsStablemaster(receiver, source game.Object, v any) bool {
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

// KeywordsCommand handles command-able creature speech triggers.
func KeywordsCommand(receiver, source game.Object, v any) bool {
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
var keywordEvents = map[string]EventHandler{
	"bank":    OpenBankBox,
	"buy":     VendorBuy,
	"claim":   ClaimAllPets,
	"come":    CommandFollowMe,
	"drop":    CommandDrop,
	"follow":  CommandFollow,
	"release": CommandRelease,
	"sell":    VendorSell,
	"stable":  StablePet,
	"stay":    CommandStay,
	"stop":    CommandStay,
}
