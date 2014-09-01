// Position support for go-toml

package toml

import (
	"fmt"
)

// position within a TOML document
type Position struct {
	Line int // line within the document
	Col  int // column within the line
}

// String representation of the position.
// Displays 1-indexed line and column numbers.
func (p *Position) String() string {
	return fmt.Sprintf("(%d, %d)", p.Line, p.Col)
}

func (p *Position) Invalid() bool {
	return p.Line <= 0 || p.Col <= 0
}
