package browser

import (
	"fmt"
	"strconv"
	"strings"
)

type VirtualID string

const VirtualIDPrefix = "vid-"
const VirtualIDDataAttr = "data-vid"

type VirtualIDGenerator interface {
	Next() VirtualID
	IsValidVirtualID(id VirtualID) bool
}

type IncrIntVirtualIDGenerator struct {
	Cur int
}

// A VirtualIDGenerator that increments an integer
func NewIncrIntVirtualIDGenerator() VirtualIDGenerator {
	return &IncrIntVirtualIDGenerator{Cur: 0}
}

func (g *IncrIntVirtualIDGenerator) Next() VirtualID {
	newID := VirtualID(fmt.Sprintf("%s%d", VirtualIDPrefix, g.Cur))
	g.Cur++
	return newID
}

func (g *IncrIntVirtualIDGenerator) IsValidVirtualID(id VirtualID) bool {
	if !IsValidBaseVirtualID(id) {
		return false
	}
	n := strings.TrimPrefix(string(id), VirtualIDPrefix)
	if _, err := strconv.Atoi(n); err != nil {
		return false
	}
	return true
}

func IsValidBaseVirtualID(id VirtualID) bool {
	return strings.HasPrefix(string(id), VirtualIDPrefix) && len(string(id)) > len(VirtualIDPrefix)
}

func VirtualIDElementQuery(id VirtualID) string {
	return fmt.Sprintf("[%s=%s]", VirtualIDDataAttr, id)
}
