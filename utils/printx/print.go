package printx

import (
	"fmt"
	"strings"
)

func PrintStandardHeader(header string) {
	hBar := strings.Repeat("-", 80)
	fmt.Println("\n" + hBar + "\n" + header + "\n" + hBar)
}
