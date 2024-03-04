// Non-allocating CSV tokenizer.
//
// CSV is not a single well-defined format.  Here is what we're parsing:
//
//  - the input is UTF-8 / ASCII, no BOM allowed (or needed).
//  - there is one CSV record per line
//  - lines are terminated by ASCII newline 0x0A, exclusively
//  - the terminator is optional at EOF
//  - blank lines are treated as empty records
//  - lines have a hardcoded max length given by MAXLINE, below
//  - there is no header line
//  - fields are separated by ASCII comma 0x2C, exclusively
//  - the number of fields can vary between the lines
//  - fields can be empty
//  - fields can be enclosed in double-quotes ASCII 0x22 and double-quotes and commas are allowed
//    inside quoted fields
//  - newlines and EOF are not allowed inside double-quoted fields
//  - a double-quote is represented as two double-quotes in a double-quoted field
//
// The tokenizer is opened on an io.Reader and provides tokens.  A token is a triplet (A, B, C)
// where A is either CsvEol, CsvEof (both of which are negative), or a nonnegative value which is
// the start index into the internal byte buffer.  If A is not CsvEol or CsvEof then B is the
// one-past-end index into the byte buffer, and C is the index within that range of the character
// following the first '=', if present, otherwise CsvEqSentinel (which is negative).  The indices
// are valid until the next call to Get() and can be passed to BufAt() and GetString() to retrieve
// contents from the internal buffer.

package sonarlog

import (
	"errors"
	"io"
)

const (
	CsvEol        = -1
	CsvEof        = -2
	CsvEqSentinel = -3
)

const (
	// The intent of the bufferSize is that sizeof(CsvTokenizer) rounds up to a convenient
	// allocation block size.  Not sure if this really matters, but we allocate one of these per
	// input stream.  TODO: It could be stack-allocated and reused.  But then, big stack-allocated
	// buffer.
	bufferSize = 65400
	// 1KB ought to be enough for anyone.
	maxLine    = 1024
)

type CsvTokenizer struct {
	reader      io.Reader
	ix          int
	lim         int
	startOfLine bool
	failRefill  bool
	buf         [bufferSize]uint8
}

func NewTokenizer(reader io.Reader) *CsvTokenizer {
	return &CsvTokenizer{
		reader:      reader,
		startOfLine: true,
	}
}

// Extract a string reference into the buffer.  start and lim must have been returned with t.get().
// This string is only valid until the next call to t.get().
func (t *CsvTokenizer) GetString(start, lim int) string {
	if start < 0 || lim < start {
		panic("Invalid GetString parameters")
	}
	return string(t.buf[start:lim])
}

// Get the byte in the buffer at the given location, this is valid exclusively for locations
// start..lim returned by a call to t.get().
func (t *CsvTokenizer) BufAt(loc int) uint8 {
	return t.buf[loc]
}

// Get the next token, or an error for syntax errors or I/O errors.
func (t *CsvTokenizer) Get() (int, int, int, error) {
	err := t.maybeRefill()
	if err != nil {
		return 0, 0, 0, err
	}

	// The following logic assumes \n is a sentinel at self.buf[self.lim].
	if t.buf[t.ix] == '\n' {
		if t.ix == t.lim {
			return CsvEof, 0, 0, nil
		}
		t.ix++
		t.startOfLine = true
		return CsvEol, 0, 0, nil
	}

	if !t.startOfLine {
		if t.buf[t.ix] != ',' {
			panic("Must have comma here")
		}
		t.ix++
	}

	// Parse a field, terminated by EOF, EOL, or comma.
	t.startOfLine = false
	switch t.buf[t.ix] {
	case '\n', ',':
		// Empty field at EOL or EOF or comma
		return t.ix, t.ix, CsvEqSentinel, nil

	case '"':
		// This is a little hairy because every doubled quote has to be collapsed into a single one.
		// We do this in the buffer, hence `destix`.
		t.ix++
		startix := t.ix
		destix := startix
		eqloc := CsvEqSentinel
		for {
			switch t.buf[t.ix] {
			case '\n':
				return 0, 0, 0, errors.New("Unexpected end of line or end of file")
			case '=':
				if eqloc == CsvEqSentinel {
					eqloc = t.ix + 1
				}
			case '"':
				t.ix++
				if t.buf[t.ix] != '"' {
					// We're done.  We've already consumed the quote.  Check that the
					// syntax is sane.
					if t.buf[t.ix] != ',' && t.buf[t.ix] != '\n' {
						return 0, 0, 0, errors.New("Expected comma or newline after quoted field")
					}
					return startix, destix, eqloc, nil
				}
			}
			t.buf[destix] = t.buf[t.ix]
			destix++
			t.ix++
		}

	default:
		startix := t.ix
		eqloc := CsvEqSentinel
		for {
			switch t.buf[t.ix] {
			case '\n', ',':
				return startix, t.ix, eqloc, nil
			case '=':
				if eqloc == CsvEqSentinel {
					eqloc = t.ix + 1
				}
			case '"':
				return 0, 0, 0, errors.New("Unexpected '\"'")
			}
			t.ix++
		}
	}
}

// Given start and non-sentinel eqloc values returned with a token and a string <tag>,
// check if the buffer has the string <tag>= from location start.
func (t *CsvTokenizer) MatchTag(tag string, start, eqloc int) bool {
	if start+len(tag)+1 != eqloc {
		return false
	}
	if t.buf[eqloc-1] != '=' {
		return false
	}
	for i := 0; i < len(tag); i++ {
		if tag[i] != t.buf[start+i] {
			return false
		}
	}
	return true
}

// Testing mode: Make the next refill fail with an error that says "Test failure"
func (t *CsvTokenizer) SetFailRefill() {
	t.failRefill = true
}

func (t *CsvTokenizer) maybeRefill() error {
	// TODO: This was a Rust-ism to fill the buffer with repeated Read calls but it may be that Go
	// is cleaner and the loop isn't required, just an `if`?
	for t.lim-t.ix < maxLine {
		if t.ix != 0 {
			n := t.lim - t.ix
			copy(t.buf[0:n], t.buf[t.ix:t.lim])
			t.ix = 0
			t.lim = n
		}
		if t.failRefill {
			// Testing mode
			return errors.New("Test failure")
		}
		nread, err := t.reader.Read(t.buf[t.lim : bufferSize-1])
		if err != nil {
			if err != io.EOF {
				return err
			}
		}
		t.lim += nread
		t.buf[t.lim] = '\n'
		if nread == 0 {
			break
		}
	}
	return nil
}
