package main

import (
	"fmt"
	"unicode/utf8"
)

type FormatOptions struct {
    Tag string // if not ""
    Json bool   // json explicitly requested
    Csv bool    // csv or csvnamed explicitly requested
    Awk bool    // awk explicitly requested
    Fixed bool  // fixed output explicitly requested
    Named bool  // csvnamed explicitly requested
    Header bool // true if nothing requested b/c fixed+header is default
    NoDefaults bool // if true and the string returned is "*skip*" and the
                      //   mode is csv or json then print nothing
}

func FormatData[T, C any](
	fields []string,
	formatters map[string]func(datum T, ctx C) string,
	opts *FormatOptions,
	data []T,
	c C,
) {
	// Produce the formatted field values for all fields
	cols := make([][]string, len(fields))
	for _, x := range data {
		for i, kwd := range fields {
			v := formatters[kwd](x, c)
			if i == 0 {
				if prefix, found := formatters["*prefix*"]; found {
					cols[i] = append(cols[i], prefix(x, c) + v)
				} else {
					cols[i] = append(cols[i], v)
				}
			} else {
				cols[i] = append(cols[i], v)
			}
		}
	}

	// Output on desired format
	if opts.Fixed {
		formatFixed(fields, opts, cols)
	} else {
		panic("Opts")
	}
}

func formatFixed(fields []string, opts *FormatOptions, cols [][]string) {
    // The column width is the max across all the entries in the column (including header,
    // if present).  If there's a tag, it is printed in the last column.
	numWidths := len(fields)
	tagCol := -1
	if opts.Tag != "" {
		tagCol = numWidths
		numWidths += 1
	}
    widths := make([]int, numWidths)

    if opts.Header {
		for col := 0 ; col < len(fields); col++ {
            widths[col] = maxi(widths[col], utf8.RuneCountInString(fields[col]))
        }
        if tagCol >= 0 {
            widths[tagCol] = maxi(widths[tagCol], len("tag"))
        }
    }

    for row := 0 ; row < len(cols[0]) ; row++ {
        for col := 0 ; col < len(fields); col++ {
            widths[col] = maxi(widths[col], utf8.RuneCountInString(cols[col][row]))
        }
        if tagCol >= 0 {
            widths[tagCol] = maxi(widths[tagCol], utf8.RuneCountInString(opts.Tag))
        }
    }

    // Header
    if opts.Header {
        s := ""
		for col := 0 ; col < len(fields); col++ {
			s += fmt.Sprintf("%*s  ", widths[col], fields[col])
        }
        if tagCol >= 0 {
			s += fmt.Sprintf("%*s  ", widths[tagCol], "tag")
        }
		println(s)
    }

    // Body
    for row := 0 ; row < len(cols[0]); row++ {
		s := ""
		for col := 0 ; col < len(fields); col++ {
			s += fmt.Sprintf("%*s  ", widths[col], cols[col][row])
        }
        if tagCol >= 0 {
			s += fmt.Sprintf("%*s  ", widths[tagCol], opts.Tag)
		}
		println(s)
    }
}

func maxi(a, b int) int {
	if a > b {
		return a
	}
	return b
}
