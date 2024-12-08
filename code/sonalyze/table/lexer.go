// See queryexpr.y for a grammar.

package table

import (
	"fmt"
)

func (p *queryParser) Lex(lval *yySymType) int {
Again:
	if p.i < len(p.input) {
		start := p.i
		c := p.input[start]
		p.i++
		switch c {
		case ' ', '\t', '\r', '\n':
			goto Again
		case '<':
			if p.i < len(p.input) && p.input[p.i] == '=' {
				p.i++
				return tLe
			}
			return tLt
		case '>':
			if p.i < len(p.input) && p.input[p.i] == '=' {
				p.i++
				return tGe
			}
			return tGt
		case '=':
			if p.i < len(p.input) && p.input[p.i] == '~' {
				p.i++
				return tMatch
			}
			return tEq
		case '(':
			return tLparen
		case ')':
			return tRparen
		case '"', '\'', '/', '`':
			for p.i < len(p.input) && p.input[p.i] != c {
				p.i++
			}
			if p.i == len(p.input) {
				p.Error(fmt.Sprintf("End of input in string at position %d", p.i-1))
				return -1
			}
			lval.text = string(p.input[start+1 : p.i])
			p.i++
			return tString
		case '-':
			if !p.scanUnsignedNumber() {
				return -1
			}
			lval.text = string(p.input[start:p.i])
			return tString
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			// Zero or more digits, so ignore result
			p.scanDigits()
			if p.i < len(p.input) && isDurationMarker(p.input[p.i]) >= 0 {
				// Easier to rewind than to hook into the middle of the logic
				p.i = start
				if !p.scanDuration() {
					return -1
				}
				lval.text = string(p.input[start:p.i])
				return tString
			}
			if !p.scanUnsignedNumberTail() {
				return -1
			}
			lval.text = string(p.input[start:p.i])
			return tString
		default:
			if isInitial(c) {
				for p.i < len(p.input) && isSubsequent(p.input[p.i]) {
					p.i++
				}
				lval.text = string(p.input[start:p.i])
				switch lval.text {
				case "and":
					return tAnd
				case "or":
					return tOr
				case "not":
					return tNot
				default:
					return tIdent
				}
			}
			p.Error(fmt.Sprintf("Unexpected character '%c' at position %d", c, p.i-1))
			return -1
		}
	}
	return -1
}

func (p *queryParser) scanUnsignedNumber() bool {
	if !p.scanNonemptyDigits() {
		return false
	}
	return p.scanUnsignedNumberTail()
}

// Initial digits have been consumed
func (p *queryParser) scanUnsignedNumberTail() bool {
	if p.i < len(p.input) && p.input[p.i] == '.' {
		p.i++
		if !p.scanNonemptyDigits() {
			return false
		}
	}
	if p.i < len(p.input) && (p.input[p.i] == 'e' || p.input[p.i] == 'E') {
		p.i++
		if p.i < len(p.input) && (p.input[p.i] == '+' || p.input[p.i] == '-') {
			p.i++
		}
		if !p.scanNonemptyDigits() {
			return false
		}
	}
	// Require a token separator after number
	if p.i < len(p.input) && isSubsequent(p.input[p.i]) {
		p.Error("Token separator required after number")
		return false
	}
	return true
}

func (p *queryParser) scanDuration() bool {
	prev := -1
	for {
		if p.i == len(p.input) || !isDigit(p.input[p.i]) {
			break
		}
		if !p.scanNonemptyDigits() {
			return false
		}
		if p.i == len(p.input) {
			p.Error("Time interval designator required in duration")
			return false
		}
		m := isDurationMarker(p.input[p.i])
		if m <= prev {
			p.Error("Time interval designators must appear in WDHM order")
			return false
		}
		p.i++
		prev = m
	}
	// Require a token separator after duration
	if p.i < len(p.input) && isSubsequent(p.input[p.i]) {
		p.Error("Token separator required after duration")
		return false
	}
	return prev > -1
}

func (p *queryParser) scanNonemptyDigits() bool {
	if !p.scanDigits() {
		p.Error("Non-empty digit string required")
		return false
	}
	return true
}

func (p *queryParser) scanDigits() bool {
	here := p.i
	for p.i < len(p.input) && isDigit(p.input[p.i]) {
		p.i++
	}
	return p.i > here
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isInitial(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_'
}

func isSubsequent(c byte) bool {
	return isInitial(c) || isDigit(c)
}

func isDurationMarker(c byte) int {
	switch c {
	case 'w', 'W':
		return 0
	case 'd', 'D':
		return 1
	case 'h', 'H':
		return 2
	case 'm', 'M':
		return 3
	default:
		return -1
	}
}
