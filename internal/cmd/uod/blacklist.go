package uod

import (
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/qbradq/sharduo/data"
	"github.com/qbradq/sharduo/lib/util"
)

// The global blacklist.
var blacklist Blacklist

func init() {
	loadBlacklist()
}

// BlacklistPatternOctet represents one octet pattern for IPv4.
type BlacklistPatternOctet uint16

const BlacklistPatternOctetAll BlacklistPatternOctet = 0xFFFF // Match any value.

// BlacklistPattern is a pattern of IPv4 address octet search parameters.
type BlacklistPattern []BlacklistPatternOctet

// NewBlacklist parses a string into a BlacklistPattern.
func NewBlacklistPattern(s string) BlacklistPattern {
	ret := make(BlacklistPattern, 4)
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return nil
	}
	for i := range parts {
		p := parts[i]
		if p == "*" {
			ret[i] = BlacklistPatternOctetAll
			continue
		}
		v, err := strconv.ParseInt(p, 0, 32)
		if err != nil {
			return nil
		}
		if v < 0 || v > 255 {
			return nil
		}
		ret[i] = BlacklistPatternOctet(v)
	}
	return ret
}

// Match returns true if there is a match between the pattern and the passed
// address.
func (p BlacklistPattern) Match(a net.IP) bool {
	for i := range p {
		if p[i] == BlacklistPatternOctetAll || p[i] == BlacklistPatternOctet(a[i]) {
			continue
		}
		return false
	}
	return true
}

// Blacklist is a collection of patterns that can be queried as a whole.
type Blacklist struct {
	p []BlacklistPattern // All of the patterns within the blacklist.
	m sync.RWMutex       // Mutex locking access to p.
}

// Match returns true if the given IP matches any of the patterns. This function
// is concurrent safe.
func (b *Blacklist) Match(a net.IP) bool {
	b.m.RLock()
	defer b.m.RUnlock()
	for _, p := range b.p {
		if p.Match(a) {
			return true
		}
	}
	return false
}

// loadBlacklist re-loads the blacklist from ./blacklist.ini .
func loadBlacklist() {
	// Make sure the file is there.
	if _, err := os.Stat("blacklist.ini"); err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
		b, err := data.FS.ReadFile(path.Join("misc", "default-blacklist.ini"))
		if err != nil {
			panic(err)
		}
		if err := os.WriteFile("blacklist.ini", b, 0660); err != nil {
			panic(err)
		}
	}
	// Start reading the list file
	var lfr util.ListFileReader
	f, err := os.Open("blacklist.ini")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	lfr.StartReading(f)
	if lfr.StreamNextSegmentHeader() != "Blacklist" {
		panic("Blacklist segment not found at top of blacklist.ini")
	}
	// Rebuild blacklist
	blacklist.m.Lock()
	defer blacklist.m.Unlock()
	blacklist.p = make([]BlacklistPattern, 0)
	for {
		s := lfr.StreamNextEntry()
		if s == "" {
			break
		}
		p := NewBlacklistPattern(s)
		if p != nil {
			blacklist.p = append(blacklist.p, p)
		}
	}
}
