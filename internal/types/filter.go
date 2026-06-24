package types

import (
	"fmt"
	"strings"
)

// FilterOp is the comparison in a single AIP-160 filter term.
type FilterOp int

const (
	// FilterEq is the `=` operator (exact match).
	FilterEq FilterOp = iota
	// FilterNeq is the `!=` operator (negated exact match).
	FilterNeq
	// FilterHas is the `:` operator (membership / substring match).
	FilterHas
)

// FilterCondition is one parsed term of a filter expression. A condition with an
// empty Field is a bareword/quoted term with no operator (e.g. `SUMMER`): a
// free-text search the adapter applies across its default searchable columns.
// Conditions in a slice are ANDed together.
type FilterCondition struct {
	Field string
	Op    FilterOp
	Value string
}

// ParseFilter parses a pragmatic subset of the AIP-160 filter language into a
// flat, AND-combined list of conditions. It supports:
//
//   - `field = value`, `field != value`, `field : value`
//   - quoted values ("two words") and barewords (SUMMER25)
//   - a bare term with no operator, treated as a free-text search
//   - terms separated by whitespace or an explicit `AND`
//
// It validates only syntax; the adapter validates each Field against the fields
// it can filter. A malformed expression yields ErrInvalidArgument. The empty
// string parses to no conditions (an unfiltered list).
func ParseFilter(filter string) ([]FilterCondition, error) {
	toks, err := tokenizeFilter(filter)
	if err != nil {
		return nil, err
	}

	var conds []FilterCondition
	for i := 0; i < len(toks); {
		t := toks[i]
		if t.op {
			return nil, fmt.Errorf("%w: unexpected operator %q in filter", ErrInvalidArgument, t.text)
		}
		if strings.EqualFold(t.text, "AND") && !t.quoted {
			i++
			continue
		}
		// A value token followed by an operator is a `field op value` term.
		if i+1 < len(toks) && toks[i+1].op {
			if i+2 >= len(toks) || toks[i+2].op {
				return nil, fmt.Errorf("%w: filter operator %q is missing a value", ErrInvalidArgument, toks[i+1].text)
			}
			op, err := filterOp(toks[i+1].text)
			if err != nil {
				return nil, err
			}
			conds = append(conds, FilterCondition{Field: t.text, Op: op, Value: toks[i+2].text})
			i += 3
			continue
		}
		// Otherwise it's a free-text term.
		conds = append(conds, FilterCondition{Op: FilterHas, Value: t.text})
		i++
	}
	return conds, nil
}

func filterOp(s string) (FilterOp, error) {
	switch s {
	case "=":
		return FilterEq, nil
	case "!=":
		return FilterNeq, nil
	case ":":
		return FilterHas, nil
	default:
		return 0, fmt.Errorf("%w: unsupported filter operator %q", ErrInvalidArgument, s)
	}
}

// filterToken is a lexed unit: a value (bareword or quoted string) or an operator.
type filterToken struct {
	text   string
	op     bool // true for =, !=, :
	quoted bool // true when the value came from a quoted string literal
}

// tokenizeFilter splits a filter expression into value and operator tokens,
// honoring double-quoted string literals (which may contain spaces and operators).
func tokenizeFilter(s string) ([]filterToken, error) {
	var toks []filterToken
	i := 0
	for i < len(s) {
		c := s[i]
		switch c {
		case ' ', '\t':
			i++
		case '"':
			j := i + 1
			var b strings.Builder
			for j < len(s) && s[j] != '"' {
				b.WriteByte(s[j])
				j++
			}
			if j >= len(s) {
				return nil, fmt.Errorf("%w: unterminated quoted string in filter", ErrInvalidArgument)
			}
			toks = append(toks, filterToken{text: b.String(), quoted: true})
			i = j + 1
		case '=', ':':
			toks = append(toks, filterToken{text: string(c), op: true})
			i++
		case '!':
			if i+1 < len(s) && s[i+1] == '=' {
				toks = append(toks, filterToken{text: "!=", op: true})
				i += 2
			} else {
				return nil, fmt.Errorf("%w: unexpected %q in filter (did you mean !=?)", ErrInvalidArgument, "!")
			}
		default:
			j := i
			for j < len(s) && !isFilterDelim(s[j]) {
				j++
			}
			toks = append(toks, filterToken{text: s[i:j]})
			i = j
		}
	}
	return toks, nil
}

func isFilterDelim(c byte) bool {
	return c == ' ' || c == '\t' || c == '=' || c == ':' || c == '!' || c == '"'
}
