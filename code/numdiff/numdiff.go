// Simple utility for approximately-equal-but-numbers-are-noisy file comparison.

// Algorithm:
// - Given two text files, read a line from both.  Fail if one is at EOF but not the other.
// - If the lines are the same, continue to next.
// - Otherwise, diff the lines char-by-char and find space-delimited corresponding fields that
//   differ. Fail if the fields are not numeric.
// - Otherwise, subtract one from the other.  Succeed if the absolute value of the difference <= 2.
// - Otherwise, compute (larger - smaller)/larger (by absolute values).  Succeed if the ratio
//   is <= 0.01.
// - Otherwise fail.

package main

import (
	"bufio"
	"flag"
	"log"
	"math"
	"os"
	"strconv"
)

const (
	absDifference = 2
	ratioDifference = 0.01
)

var comma = flag.Bool("comma", false, "Use comma separation (quasi-CSV)")

func main() {
	flag.Parse()
	rest := flag.Args()
	if len(rest) != 2 {
		log.Fatalf("Usage: %s [-comma] input1 input2", os.Args[0])
	}
	f1, err := os.Open(rest[0])
	if err != nil {
		log.Fatal(err)
	}
	defer f1.Close()
	f2, err := os.Open(rest[1])
	if err != nil {
		log.Fatal(err)
	}
	defer f2.Close()
	s1 := bufio.NewScanner(f1)
	s2 := bufio.NewScanner(f2)
	lineno := 0
LineLoop:
	for {
		r1 := s1.Scan()
		r2 := s2.Scan()
		if !r1 && !r2 {
			break LineLoop
		}
		lineno++
		if r1 != r2 {
			log.Fatalf("Line %d: Files of different length", lineno)
		}
		t1 := s1.Text()
		t2 := s2.Text()
		if t1 == t2 {
			continue LineLoop
		}
		fld1 := split(t1)
		fld2 := split(t2)
		if len(fld1) != len(fld2) {
			log.Fatalf("Line %d: Lines with different numbers of fields", lineno)
		}
	FieldLoop:
		for fldIx := 0; fldIx < len(fld1); fldIx++ {
			if fld1[fldIx] == fld2[fldIx] {
				continue FieldLoop
			}
			a := fld1[fldIx]
			b := fld2[fldIx]
			na, err := strconv.ParseFloat(a, 64)
			if err != nil {
				log.Fatalf("Line %d: Non-numeric diverging field `%s`", lineno, a)
			}
			na = math.Abs(na)
			nb, err := strconv.ParseFloat(b, 64)
			if err != nil {
				log.Fatalf("Line %d: Non-numeric diverging field `%s`", lineno, b)
			}
			nb = math.Abs(nb)
			larger := math.Max(na, nb)
			smaller := math.Min(na, nb)
			if larger-smaller <= absDifference {
				continue FieldLoop
			}
			if (larger-smaller)/larger <= ratioDifference {
				continue FieldLoop
			}
			log.Fatalf("Line %d: Fields differ too much: `%s` `%s`", lineno, a, b)
		}
	}
}

// Split into runs of spaces and non-spaces.
func split(s string) []string {
	ss := make([]string, 0)
	i := 0
	if *comma {
		for lim, x := len(s), true ; i < lim ; x = !x {
			start := i
			if i < lim && (s[i] == ',') == x {
				i++
			}
			if start != i {
				ss = append(ss, s[start:i])
			}
		}
	} else {
		for lim, x := len(s), true ; i < lim ; x = !x {
			start := i
			for i < lim && (s[i] == ' ' || s[i] == '\t') == x {
				i++
			}
			if start != i {
				ss = append(ss, s[start:i])
			}
		}
	}
	return ss
}
