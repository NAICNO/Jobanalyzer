package sonarlog

import (
	"bytes"
	"testing"
)

// eqloc is either CsvEqSentinel or it's a delta from a
func checkMatch(t *testing.T, tokenizer *CsvTokenizer, s string, eqloc int) {
	a, b, c, err := tokenizer.Get()
	if err != nil {
		t.Fatal(err)
	}
	if a < 0 {
		t.Fatalf("Unexpected start value %d", a)
	}
	s2 := tokenizer.GetString(a, b)
	if s2 != s {
		t.Fatalf("Not a match for value '%s': '%s'", s, s2)
	}
	if (eqloc == CsvEqSentinel && c != eqloc) || (eqloc != CsvEqSentinel && c != a+eqloc) {
		t.Fatalf("Not a match for eqloc %d: %d", eqloc, c-a)
	}
}

// eqloc is either CsvEqSentinel or it's a delta from a
func checkSentinel(t *testing.T, tokenizer *CsvTokenizer, v int) {
	a, _, _, err := tokenizer.Get()
	if err != nil {
		t.Fatal(err)
	}
	if a != v {
		t.Fatalf("Unexpected sentinel value %d", a)
	}
}

func checkErr(t *testing.T, tokenizer *CsvTokenizer) {
	_, _, _, err := tokenizer.Get()
	if err == nil {
		t.Fatal("Expected error")
	}
}

// This tests:
//  - empty fields, also at EOL (but not at EOF)
//  - quoted fields, also with commas and quotes in them
//  - unterminated last line
//  - blank line / empty record

func TestCsvTokenizer1(t *testing.T) {
	text := `a,b=1,cc=2,,e,"f=1,2,3","g,""y"",z",

A,B`
	tokenizer := NewTokenizer(bytes.NewReader([]byte(text)))
	checkMatch(t, tokenizer, "a", CsvEqSentinel)
	checkMatch(t, tokenizer, "b=1", 2)
	checkMatch(t, tokenizer, "cc=2", 3)
	checkMatch(t, tokenizer, "", CsvEqSentinel)
	checkMatch(t, tokenizer, "e", CsvEqSentinel)
	checkMatch(t, tokenizer, "f=1,2,3", 2)
	checkMatch(t, tokenizer, "g,\"y\",z", CsvEqSentinel)
	checkMatch(t, tokenizer, "", CsvEqSentinel)
	checkSentinel(t, tokenizer, CsvEol)
	checkSentinel(t, tokenizer, CsvEol)
	checkMatch(t, tokenizer, "A", CsvEqSentinel)
	checkMatch(t, tokenizer, "B", CsvEqSentinel)
	checkSentinel(t, tokenizer, CsvEof)
}

// This tests:
//  - empty field at EOF

func TestCsvTokenizer2(t *testing.T) {
	text := `a,`
	tokenizer := NewTokenizer(bytes.NewReader([]byte(text)))
	checkMatch(t, tokenizer, "a", CsvEqSentinel)
	checkMatch(t, tokenizer, "", CsvEqSentinel)
	checkSentinel(t, tokenizer, CsvEof)
}

// This tests:
//  - syntax error: eol in quoted string

func TestCsvTokenizer3(t *testing.T) {
	text := `a,"hi
ho`
	tokenizer := NewTokenizer(bytes.NewReader([]byte(text)))
	checkMatch(t, tokenizer, "a", CsvEqSentinel)
	checkErr(t, tokenizer)
}

// This tests:
//  - syntax error: eof in quoted string

func TestCsvTokenizer4(t *testing.T) {
	text := `a,"hi`
	tokenizer := NewTokenizer(bytes.NewReader([]byte(text)))
	checkMatch(t, tokenizer, "a", CsvEqSentinel)
	checkErr(t, tokenizer)
}

// This tests:
//  - syntax error: junk following quoted string

func TestCsvTokenizer5(t *testing.T) {
	text := `a,"hi"x,y`
	tokenizer := NewTokenizer(bytes.NewReader([]byte(text)))
	checkMatch(t, tokenizer, "a", CsvEqSentinel)
	checkErr(t, tokenizer)
}

// This tests:
//  - syntax error: quote in unquoted string

func TestCsvTokenizer6(t *testing.T) {
	text := `a,hi"x,y`
	tokenizer := NewTokenizer(bytes.NewReader([]byte(text)))
	checkMatch(t, tokenizer, "a", CsvEqSentinel)
	checkErr(t, tokenizer)
}

// This tests:
//  - refill logic
//
// Basically we're creating an input s.t. the characters for a token overlap the buffer boundary.
// Parts of the token are read on the initial fill, then the rest on the second fill.  The file
// contains single-token lines, each line is the 26-character string a...z followed by a newline.
// The test then just checks that every token looks right.

func TestCsvTokenizer7(t *testing.T) {
	// Token+newline must straddle the buffer boundary in some interesting way
	if !(bufferSize%27 != 0 && bufferSize%27 != 1 && bufferSize%27 != 26) {
		panic("Failed precondition")
	}
	text := ""
	count := bufferSize * 3 / 27
	for i := 0; i < count; i++ {
		text += "abcdefghijklmnopqrstuvwxyz\n"
	}
	tokenizer := NewTokenizer(bytes.NewReader([]byte(text)))
	state := 0
	found := 0
Loop:
	for {
		a, b, c, err := tokenizer.Get()
		if err != nil {
			t.Fatal(err)
		}
		switch a {
		case CsvEol:
			if state != 1 {
				t.Fatal("Bad state at EOL")
			}
			state = 0
		case CsvEof:
			if state != 0 {
				t.Fatal("Bad state at EOF")
			}
			break Loop
		default:
			if state != 0 {
				t.Fatal("Bad state at token")
			}
			if tokenizer.GetString(a, b) != "abcdefghijklmnopqrstuvwxyz" {
				t.Fatal("Bad string")
			}
			if c != CsvEqSentinel {
				t.Fatal("Bad sentinel")
			}
			state = 1
			found++
		}
	}
	if found != count {
		t.Fatal("Did not find the expected count")
	}
}

// This tests:
//  - i/o error on refill

func TestCsvTokenizer8(t *testing.T) {
	// This is the same input as test_csv_tokenizer7() for a reason, see below.
	text := ""
	count := bufferSize * 3 / 27
	for i := 0; i < count; i++ {
		text += "abcdefghijklmnopqrstuvwxyz\n"
	}
	tokenizer := NewTokenizer(bytes.NewReader([]byte(text)))
	_, _, _, err := tokenizer.Get() // Fill once
	if err != nil {
		panic(err)
	}
	tokenizer.SetFailRefill()
	for {
		a, _, _, err := tokenizer.Get()
		if err != nil {
			// This can only be an I/O error because TestCsvTokenizer7() would otherwise have
			// encountered it, since it uses the same input.
			break
		}
		if a == CsvEof {
			t.Fatal("EOF")
		}
	}
}

func checkMatchTag(t *testing.T, tokenizer *CsvTokenizer, tag, s string) {
	a, b, c, err := tokenizer.Get()
	if err != nil {
		t.Fatal(err)
	}
	if a < 0 {
		t.Fatalf("Unexpected start value %d", a)
	}
	if c == CsvEqSentinel {
		t.Fatal("Expected eqloc")
	}
	if !tokenizer.MatchTag(tag, a, c) {
		t.Fatal("MatchTag should succeed")
	}
	if tokenizer.MatchTag("nih", a, c) {
		t.Fatal("MatchTag should fail")
	}
	s2 := tokenizer.GetString(c, b)
	if s2 != s {
		t.Fatalf("Not a match for value '%s': '%s'", s, s2)
	}
}

// Test the MatchTag method

func TestCsvTokenizer9(t *testing.T) {
	text := `a=10,baba=20,cc=2,dd=`
	tokenizer := NewTokenizer(bytes.NewReader([]byte(text)))
	checkMatchTag(t, tokenizer, "a", "10")
	checkMatchTag(t, tokenizer, "baba", "20")
	checkMatchTag(t, tokenizer, "cc", "2")
	checkMatchTag(t, tokenizer, "dd", "")
	checkSentinel(t, tokenizer, CsvEof)
}
