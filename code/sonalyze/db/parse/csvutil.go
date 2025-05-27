// Various utilities used to handle the output of the fast CSV parser

package parse

import (
	"bytes"
	"errors"
	"math"
	"time"
)

func match(tokenizer *CsvTokenizer, start, lim, eqloc int, tag string) ([]byte, bool) {
	if tokenizer.MatchTag(tag, start, eqloc) {
		return tokenizer.BufSlice(eqloc, lim), true
	}
	return nil, false
}

// Old name, always meant uint64
func parseUint(bs []byte) (uint64, error) {
	return parseUint64(bs)
}

func parseUint64(bs []byte) (uint64, error) {
	var n uint64
	if len(bs) == 0 {
		return 0, errors.New("Empty")
	}
	for _, c := range bs {
		if c < '0' || c > '9' {
			return 0, errors.New("Not a digit")
		}
		m := n*10 + uint64(c-'0')
		if m < n {
			return 0, errors.New("Out of range")
		}
		n = m
	}
	return n, nil
}

func parseUint32(bs []byte) (uint32, error) {
	n, err := parseUint64(bs)
	if err != nil {
		return 0, err
	}
	if n > math.MaxUint32 {
		return 0, errors.New("Overflow")
	}
	return uint32(n), nil
}

func parseUint8(bs []byte) (uint8, error) {
	n, err := parseUint64(bs)
	if err != nil {
		return 0, err
	}
	if n > math.MaxUint8 {
		return 0, errors.New("Overflow")
	}
	return uint8(n), nil
}

// A faster number parser operating on byte slices.
//
// This is primitive and - except for NaN and Infinity - handles only simple unsigned numbers with a
// fraction, no exponent.  Accuracy is not great either.  But it's good enough for the Sonar output,
// which should have no exponentials, require low accuracy, and only occasionally - in older,
// buggier data - has NaN and Infinity.
//
// According to documentation, strconv.ParseFloat() accepts nan, inf, +inf, -inf, infinity,
// +infinity and -infinity, case-insensitively.
//
// Based on experimentation, the rust to_string() formatter will produce "NaN", "inf" and "-inf",
// with that capitalization.
//
// Based on experimentation, the Go formatter produces "NaN", "+Inf" and "-Inf".
func parseFloat(bs []byte, filterInfNaN bool) (float64, error) {
	// Canonical code
	// x, err := strconv.ParseFloat(string(bs), 64)
	// if err != nil {
	// 	return 0, err
	// }
	// if filterInfNaN && (math.IsInf(x, 0) || math.IsNaN(x)) {
	// 	return 0, errors.New("Infinity / NaN")
	// }
	// return x, nil
	var n float64
	if len(bs) == 0 {
		return 0, errors.New("Empty")
	}
	switch bs[0] {
	case '-':
		// No negative numbers
		return 0, errors.New("Not a digit")
	case '+':
		if bytes.EqualFold(bs, []byte{'+', 'i', 'n', 'f', 'i', 'n', 'i', 't', 'y'}) ||
			bytes.EqualFold(bs, []byte{'+', 'i', 'n', 'f'}) {
			if filterInfNaN {
				return 0, errors.New("Infinity")
			}
			return math.Inf(1), nil
		}
		return 0, errors.New("Not a digit")
	case 'i', 'I':
		if bytes.EqualFold(bs, []byte{'i', 'n', 'f', 'i', 'n', 'i', 't', 'y'}) ||
			bytes.EqualFold(bs, []byte{'i', 'n', 'f'}) {
			if filterInfNaN {
				return 0, errors.New("Infinity")
			}
			return math.Inf(1), nil
		}
		return 0, errors.New("Not a digit")
	case 'n', 'N':
		if bytes.EqualFold(bs, []byte{'n', 'a', 'n'}) {
			if filterInfNaN {
				return 0, errors.New("NaN")
			}
			return math.NaN(), nil
		}
		return 0, errors.New("Not a digit")
	}
	i := 0
	for ; i < len(bs); i++ {
		c := bs[i]
		if c == '.' {
			break
		}
		if c < '0' || c > '9' {
			return 0, errors.New("Not a digit")
		}
		n = n*10 + float64(c-'0')
	}
	if i < len(bs) {
		if bs[i] != '.' {
			return 0, errors.New("Only decimal point allowed")
		}
		i++
		if i == len(bs) {
			return 0, errors.New("Empty fraction")
		}
		f := 0.1
		for ; i < len(bs); i++ {
			c := bs[i]
			if c < '0' || c > '9' {
				return 0, errors.New("Not a digit")
			}
			n += float64(c-'0') * f
			f *= 0.1
		}
	}
	return n, nil
}

// The format is supposedly [DD-[HH:]]MM:SS[.micros] though some parts of the Slurm doc say that the
// HH: are not optional, and some parts do not specify directly that these are two-digit values.  So
// probably best to just look for terminators and parse the numbers between.
//
// This code is performance-sensitive, we'll use this multiple times for every record read.
//
// We'll simplify the grammar: (DD-)?(HH:)?MM:SS(.micros)?

var (
	SlurmElapsedError = errors.New("Bad elapsed time format")
)

func parseSlurmElapsed64(bs []byte) (uint64, error) {
	var days, hours, minutes, seconds uint64
	var haveDays, haveHours, haveMinutes bool

	n, i := parseUint64Here(bs, 0)
	if i < 0 {
		return 0, SlurmElapsedError
	}

	// Value followed by -
	if i < len(bs) && bs[i] == '-' {
		days = n
		haveDays = true
		i++
		n, i = parseUint64Here(bs, i)
		if i < 0 {
			return 0, SlurmElapsedError
		}
	}

	// Values followed by colon, this could be HH or MM
	for i < len(bs) && bs[i] == ':' {
		i++
		if haveHours {
			return 0, SlurmElapsedError
		}
		if haveMinutes {
			hours = minutes
			haveHours = true
		}
		minutes = n
		haveMinutes = true
		n, i = parseUint64Here(bs, i)
		if i < 0 {
			return 0, SlurmElapsedError
		}
	}

	if !haveMinutes {
		return 0, SlurmElapsedError
	}

	// Value not followed by colon
	seconds = n

	// Chop off the micros
	if i < len(bs) && bs[i] == '.' {
		i++
		n, i = parseUint64Here(bs, i)
		if i < 0 {
			return 0, SlurmElapsedError
		}
	}

	// Better be done
	if i < len(bs) {
		return 0, SlurmElapsedError
	}

	result := seconds + minutes*60
	if haveHours {
		result += hours * 3600
	}

	if haveDays {
		result += days * 3600 * 24
	}

	return result, nil
}

func parseSlurmElapsed32(bs []byte) (uint32, error) {
	n, err := parseSlurmElapsed64(bs)
	if err != nil {
		return 0, err
	}
	if n > math.MaxUint32 {
		return 0, SlurmElapsedError
	}
	return uint32(n), nil
}

func parseUint64Here(bs []byte, i int) (uint64, int) {
	start := i
	var n uint64
	for i < len(bs) && bs[i] >= '0' && bs[i] <= '9' {
		old := n
		n = n*10 + uint64(bs[i]-'0')
		if n < old {
			return 0, -1
		}
		i++
	}
	if i == start {
		return 0, -1
	}
	return n, i
}

// yyyy-mm-ddThh:mm:ssZ, yyyy-mm-ddThh:mm:ss+hh:mm, etc; no nano part, but time zone is required

func parseRFC3339(bs []byte) (t int64, err error) {
	var tmp time.Time
	tmp, err = time.Parse(time.RFC3339, string(bs))
	t = tmp.Unix()
	return
}

// The format is an integer or floating point value possibly followed by K, M, G.  The
// result is always in gigabytes, always rounded up.

func parseSlurmBytes(bs []byte) (bytes uint32, err error) {
	mpy := float64(1)
	switch bs[len(bs)-1] {
	case 'G':
		mpy = 1024 * 1024 * 1024
		bs = bs[:len(bs)-1]
	case 'M':
		mpy = 1024 * 1024
		bs = bs[:len(bs)-1]
	case 'K':
		mpy = 1024
		bs = bs[:len(bs)-1]
	}
	n, err := parseFloat(bs, true)
	if err != nil {
		return 0, err
	}
	return uint32(math.Ceil((n * mpy) / (1024 * 1024 * 1024))), nil
}
