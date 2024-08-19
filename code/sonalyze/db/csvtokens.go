// Non-allocating CSV tokenizer.
//
// CSV is not a single well-defined format.  Here is what we're parsing:
//
//  - the input is UTF-8 / ASCII, no BOM allowed (or needed).
//  - there is one CSV record per line
//  - lines are terminated by ASCII newline 0x0A, exclusively
//  - the terminator is optional at EOF
//  - blank lines are treated as empty records (not ignored)
//  - lines have a hardcoded max length given by `maxLine`, below
//  - there is no header line
//  - fields are separated by ASCII comma 0x2C, exclusively
//  - the number of fields can vary from line to line
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
// following the first '=', if present and not at the initial location, otherwise CsvEqSentinel
// (which is negative).  The indices are valid until the next call to Get() and can be passed to
// BufAt(), BufSlice() and BufSubstringSlow() to retrieve contents from the internal buffer.

package db

import (
	"errors"
	"fmt"
	"io"
)

const (
	CsvEol        = -1
	CsvEof        = -2
	CsvEqSentinel = -3
)

const (
	// The intent of the bufferSize is that sizeof(CsvTokenizer) rounds up to a convenient
	// allocation block size and is kind to the allocator.  I'm not sure if this really matters, but
	// we do allocate one of these per input stream.  If allocating a lot of these becomes an issue
	// we could use a sync.Pool to manage and reuse them.
	bufferSize = 65400

	// 1KB ought to be enough for anyone.  The buffer is backfilled if the amount of unread data
	// falls below this.
	maxLine = 1024
)

type CsvTokenizer struct {
	reader      io.Reader
	ix          int
	lim         int
	startOfLine bool
	failRefill  bool
	buf         [bufferSize]uint8
}

// MT: Constant after initialization; immutable
var SyntaxErr = errors.New("CSV syntax error")

func NewTokenizer(reader io.Reader) *CsvTokenizer {
	return &CsvTokenizer{
		reader:      reader,
		startOfLine: true,
	}
}

// Get the byte in the buffer at the given location.  `loc` must be in the range [start,lim) as
// returned by a call to t.Get().
func (t *CsvTokenizer) BufAt(loc int) uint8 {
	return t.buf[loc]
}

// Make a slice from part of the buffer.  start and lim must have been returned from t.Get(), no
// calls to t.Get() must have been made since the values were obtained, and the slice is valid only
// until the next call to t.Get().
func (t *CsvTokenizer) BufSlice(start, lim int) []byte {
	return t.buf[start:lim]
}

// Make a string from part of the buffer (this allocates a fresh string).  start and lim must have
// been returned from t.Get(), and no calls to t.Get() must have been made since the values were
// obtained.
func (t *CsvTokenizer) BufSubstringSlow(start, lim int) string {
	if start < 0 || lim < start {
		panic("Invalid GetString parameters")
	}
	return string(t.buf[start:lim])
}

// Get the next token, or an error for syntax errors or I/O errors.  Syntax errors wrap SyntaxErr,
// above.  Every other error is an I/O error.
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
			panic("Inconsistent: Must have comma here")
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
		// We do this in the buffer, hence `destix`.  We could optimize this in the same way as the
		// nonquoted scanning loop below but it's probably not worth the bother because quoted
		// fields are uncommon in our data.
		t.ix++
		startix := t.ix
		destix := startix
		eqloc := CsvEqSentinel
		for {
			switch t.buf[t.ix] {
			case '\n':
				return 0, 0, 0, fmt.Errorf("%w: Unexpected end of line or end of file", SyntaxErr)
			case '=':
				if t.ix-startix > 0 && eqloc == CsvEqSentinel {
					eqloc = t.ix + 1
				}
			case '"':
				t.ix++
				if t.buf[t.ix] != '"' {
					// We're done.  We've already consumed the quote.  Check that the
					// syntax is sane.
					if t.buf[t.ix] != ',' && t.buf[t.ix] != '\n' {
						return 0, 0, 0, fmt.Errorf(
							"%w: Expected comma or newline after quoted field",
							SyntaxErr,
						)
					}
					return startix, destix, eqloc, nil
				}
			}
			t.buf[destix] = t.buf[t.ix]
			destix++
			t.ix++
		}

	default:
		// The scanning loop is very hot so let's optimize by simplifying it and hoisting the index.
		// But hoisting buf is (so far) not helpful.
		ix := t.ix
		startix := ix
		eqloc := CsvEqSentinel
		for {
			for t.buf[ix] != '\n' && t.buf[ix] != ',' && t.buf[ix] != '"' && t.buf[ix] != '=' {
				ix++
			}
			if t.buf[ix] == '=' {
				if ix-startix > 0 && eqloc == CsvEqSentinel {
					eqloc = ix + 1
				}
				ix++
				continue
			}
			t.ix = ix
			if t.buf[ix] == '"' {
				return 0, 0, 0, fmt.Errorf("%w: Unexpected '\"'", SyntaxErr)
			}
			return startix, ix, eqloc, nil
		}
	}
}

// Fast-forward until we find an EOL (or an end-of-buffer sentinel).  Use this to discard input upon
// discovering an error.
func (t *CsvTokenizer) ScanEol() {
	for t.buf[t.ix] != '\n' {
		t.ix++
	}
	if t.ix != t.lim {
		t.ix++
	}
	t.startOfLine = true
}

// Given `start` and non-sentinel `eqloc` values returned with a token and a string `tag`, check if
// the buffer has the string `tag=` from location `start`.  Note there's no need to check that we're
// in-bounds since we're allowed to assume that `eqloc` is legitimate.
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
	// Read until we've got at least `maxLine` bytes to read.
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
		// The io.Reader contract is a little hairy:
		//
		//   When Read encounters an error or end-of-file condition after successfully reading n > 0
		//   bytes, it returns the number of bytes read. It may return the (non-nil) error from the
		//   same call or return the error (and n == 0) from a subsequent call. An instance of this
		//   general case is that a Reader returning a non-zero number of bytes at the end of the
		//   input stream may return either err == EOF or err == nil. The next Read should return 0,
		//   EOF.
		//
		//   Callers should always process the n > 0 bytes returned before considering the error err.
		t.lim += nread
		t.buf[t.lim] = '\n'
		if err != nil {
			if err == io.EOF {
				break
			}
			// I guess we may have gotten some input to process too but there seems to be no shame
			// in just returning the error at this point.
			return err
		}
		if nread == 0 {
			break
		}
	}
	return nil
}
