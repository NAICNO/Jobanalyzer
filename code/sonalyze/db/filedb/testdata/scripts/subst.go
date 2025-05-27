///usr/bin/env go run "$0" "$@"; exit "$?"

// subst is a simple substituter with a little intelligence.  Input is read from stdin and written
// to stdout.  The only, and required, program argument is the name of a pattern file.  The pattern
// file has lines of two forms (other lines are ignored):
//
// S old new
// C left-context regexp prefix
//
// An S line replaces old text with new text.  A C line finds the left-context, then matches the
// regexp at the point following the context, and if it matches, replaces the match with a string
// generated from the given prefix and a serial number.  Equal matching strings will have equal
// substitutions.
//
// PATTERNS ARE APPLIED LINE-WISE WITH ALL S PATTERNS IN ORDER FIRST FOLLOWED BY ALL C PATTERNS IN
// ORDER.  Buyer beware.  Each pattern is applied to the result of the previous substitution.
//
// None of the fields can contain spaces.
package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Missing pattern file")
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	run(f, os.Stdin, os.Stdout)
}

func run(patterns, input io.Reader, output io.Writer) {
	type S struct {
		oldText string
		newText string
	}
	simple := make([]S, 0)

	type C struct {
		context string
		matcher *regexp.Regexp
		prefix  string
	}
	contextual := make([]C, 0)

	scanner := bufio.NewScanner(patterns)
	for scanner.Scan() {
		l := scanner.Text()
		var a, b, c string
		if n, _ := fmt.Sscanf(l, "S %s %s", &a, &b); n == 2 {
			simple = append(simple, S{a, b})
			continue
		}
		if n, _ := fmt.Sscanf(l, "C %s %s %s", &a, &b, &c); n == 3 {
			contextual = append(contextual, C{a, regexp.MustCompile("^(" + b + ")"), c})
			continue
		}
	}

	const bufsiz = 2*1024*1024

	scanner = bufio.NewScanner(input)
	scanner.Buffer(make([]byte, bufsiz), bufsiz)
	for scanner.Scan() {
		l := scanner.Text()
		for _, s := range simple {
			l = strings.ReplaceAll(l, s.oldText, s.newText)
		}
		for _, c := range contextual {
			l = substitute(l, c.context, c.matcher, c.prefix)
		}
		fmt.Fprintln(output, l)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func substitute(l, leftCtx string, matcher *regexp.Regexp, prefix string) string {
	ix := strings.Index(l, leftCtx)
	if ix == -1 {
		return l
	}
	start := ix + len(leftCtx)
	m := matcher.FindStringSubmatch(l[start:])
	if m == nil {
		return l[:ix+1] + substitute(l[ix+1:], leftCtx, matcher, prefix)
	}
	return l[:start] + mapping(m[1], prefix) + substitute(l[start+len(m[1]):], leftCtx, matcher, prefix)
}

var (
	mappings = make(map[string]string)
	next     = 10000
)

func mapping(name, prefix string) string {
	if subst, found := mappings[name]; found {
		return subst
	}
	newName := fmt.Sprintf("%s%d", prefix, next)
	next++
	mappings[name] = newName
	return newName
}
