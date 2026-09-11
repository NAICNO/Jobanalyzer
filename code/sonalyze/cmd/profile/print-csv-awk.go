// Output format:
//
// For CSV and AWK output the "fields" list must have a single string from the set
// "cpu","mem","res","gpu","gpumem".  This is the per-process value we print, along with a
// timestamp.  But the layout differs among these three formats:
//
//  - For csv, each row has a timestamp and then one data field for each process at that time, that
//    is, time increases along the y axis.  A header row is printed, with "time" in the first field
//    and process information (command and pid) in each subsequent field.  There will be empty
//    fields where there are no data for a process at a time.
//
//  - For awk, the format should be as for csv except the field separator is a space, as normal.

package profile

import (
	"fmt"
	"io"

	"sonalyze/data/sample"
	. "sonalyze/table"
)

func (pc *ProfileCommand) printCsvOrAwk(
	out io.Writer,
	m *profData,
	processes []sample.SampleStream,
	pif *processIndexFactory,
) error {
	header, matrix, err := pc.collectCsvOrAwk(m, processes, pif)
	if err != nil {
		return err
	}
	if pc.PrintOpts.Csv {
		FormatRawRowmajorCsv(out, header, matrix)
	} else {
		FormatRawRowmajorAwk(out, header, matrix)
	}
	return nil
}

// This returns a row-of-columns representation, and the header will be nil if a header is not
// explicitly requested in the options.
//
// NOTE, for multi-host jobs the host is encoded in the header names, this is a tricky matter and
// really not a super happy outcome (esp for awk, since the header is not printed by default).  It's
// no worse than the PID or command, really, but the header field is becoming dangerously
// overloaded.  The syntax of each header field is always <something>(pid@host) or <something>(pid)
// so it is not a crisis, but we could consider cleaning this up.

func (pc *ProfileCommand) collectCsvOrAwk(
	m *profData,
	processes []sample.SampleStream,
	pif *processIndexFactory,
) (header []string, matrix [][]string, err error) {
	var formatter func(*profDatum) string
	formatter, err = lookupSingleFormatter(pc.PrintFields, false)
	if err != nil {
		return
	}

	if pc.PrintOpts.Header {
		header = []string{"time"}
		sep := ""
		if pc.PrintOpts.Csv {
			sep = " "
		}
		for _, process := range processes {
			pid := pif.indexFor(process[0])
			header = append(header,
				fmt.Sprintf(
					"%s%s(%s)", process[0].Cmd, sep, pif.nameFor(pid)))
		}
	}

	// Iterate by processes rather than m.cols() since process order is what we care about.

	rowNames := m.rows()
	matrix = make([][]string, len(rowNames))
	for i := range matrix {
		matrix[i] = make([]string, len(processes)+1)
	}

	for y, rn := range rowNames {
		matrix[y][0] = formatTime(rn)
		for x, p := range processes {
			pid := pif.indexFor(p[0])
			entry := m.get(rn, pid)
			if entry != nil {
				matrix[y][x+1] = formatter(entry)
			}
		}
	}

	return
}
