package virtualid

import (
	"fmt"
	"strconv"
	"strings"
)

const VirtualIDPrefix = "vid-"
const VirtualIDDataAttr = "data-vid"

type VirtualIDGenerator interface {
	Generate() string
	IsValidVirtualID(id string) bool
}

type IncrIntVirtualIDGenerator struct {
	Cur int
}

// A VirtualIDGenerator that increments an integer
func NewIncrIntVirtualIDGenerator() VirtualIDGenerator {
	return &IncrIntVirtualIDGenerator{Cur: 0}
}

func (g *IncrIntVirtualIDGenerator) Generate() string {
	newID := fmt.Sprintf("%s%d", VirtualIDPrefix, g.Cur)
	g.Cur++
	return newID
}

func (g *IncrIntVirtualIDGenerator) IsValidVirtualID(id string) bool {
	if !IsValidBaseVirtualID(id) {
		return false
	}
	n := strings.TrimPrefix(string(id), VirtualIDPrefix)
	if _, err := strconv.Atoi(n); err != nil {
		return false
	}
	return true
}

func IsValidBaseVirtualID(id string) bool {
	return strings.HasPrefix(string(id), VirtualIDPrefix) && len(string(id)) > len(VirtualIDPrefix)
}

func VirtualIDElementQuery(id string) string {
	return fmt.Sprintf(`[%s="%s"]`, VirtualIDDataAttr, id)
}
