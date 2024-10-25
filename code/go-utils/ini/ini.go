// Parse very simple ini files.
//
// The file format is line-oriented.
//
// The file is first stripped of comment lines and blank lines:
//   COMMENT = /^\s*#.*$/
//   BLANK = /^\s*$/
//
// The remaining file must then conform to this grammar:
//   file ::= section*
//   section ::= section-header section-statement*
//   section-header ::= /^\[IDENT\]\s*$/
//   section-statement ::= /^\s*IDENT\s*=VALUE$/
//
// where
//   IDENT = /[-a-zA-Z_$][-a-zA-Z0-9_$]*/
//   VALUE = .*

package ini

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
)

type IniFile map[string]*IniSection

type IniSection struct {
	Name string
	Vars map[string]string
}

func (ini IniSection) String() string {
	s := fmt.Sprintf("[%s]\n", ini.Name)
	for name, val := range ini.Vars {
		s += fmt.Sprintf("%s=%s\n", name, val)
	}
	return s
}

var (
	commentOrBlankLine = regexp.MustCompile(`^\s*(#.*)?$`)
	ident              = `[-a-zA-Z_$][-a-zA-Z0-9_$]*`
	headerLine         = regexp.MustCompile(`^\[(` + ident + `)\]\s*$`)
	sectionStmtLine    = regexp.MustCompile(`^\s*(` + ident + `)\s*=(.*)$`)
)

// This will error out on anything malformed or on duplicated section headers, the error message
// will contain line number and text.

func ParseIni(input io.Reader) (ini IniFile, err error) {
	ini = make(map[string]*IniSection)
	lineNo := 0
	scanner := bufio.NewScanner(input)
	var currentSection *IniSection
	for scanner.Scan() {
		l := scanner.Text()
		lineNo++
		if commentOrBlankLine.MatchString(l) {
			continue
		}
		if m := headerLine.FindStringSubmatch(l); m != nil {
			name := m[1]
			if _, found := ini[name]; found {
				err = fmt.Errorf("Line %d: Duplicated section name %s.\n%s", lineNo, name, l)
				return
			}
			currentSection = &IniSection{Name: name, Vars: make(map[string]string)}
			ini[name] = currentSection
			continue
		}
		if currentSection == nil {
			err = fmt.Errorf("Line %d: Missing section header\n%s", lineNo, l)
			return
		}
		if m := sectionStmtLine.FindStringSubmatch(l); m != nil {
			name := m[1]
			value := m[2]
			if _, found := currentSection.Vars[name]; found {
				err = fmt.Errorf("Line %d: Duplicated variable name %s.\n%s", lineNo, name, l)
				return
			}
			currentSection.Vars[name] = value
			continue
		}
		err = fmt.Errorf("Line %d: Malformed content.\n%s", lineNo, l)
		return
	}
	return
}
