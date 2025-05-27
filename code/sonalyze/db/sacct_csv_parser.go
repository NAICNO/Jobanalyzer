// Parser for sacct data.  This is optimized; it uses the same structure as the Sample parser.

package db

import (
	"bytes"
	"errors"
	"io"
	"math"

	. "sonalyze/common"
)

func ParseSlurmCSV(
	input io.Reader,
	ustrs UstrAllocator,
	verbose bool,
) (
	records []*SacctInfo,
	softErrors int,
	err error,
) {
	records = make([]*SacctInfo, 0)
	tokenizer := NewTokenizer(input)
	endOfInput := false
	gputmp := make([]byte, 100)

LineLoop:
	for !endOfInput {
		anyMatched := false
		info := &SacctInfo{
			End: math.MaxInt64,
		}

	FieldLoop:
		for {
			var start, lim, eqloc int
			var matched bool
			start, lim, eqloc, err = tokenizer.Get()
			if err != nil {
				if !errors.Is(err, SyntaxErr) {
					return
				}
				tokenizer.ScanEol()
				softErrors++
				continue LineLoop
			}

			if start == CsvEol {
				break FieldLoop
			}

			if start == CsvEof {
				endOfInput = true
				break FieldLoop
			}

			// NOTE, in error cases below we don't extract the offending field b/c it seems the
			// optimizer will hoist the (technically effect-free) extraction out of the parsing
			// switch and slow everything down tremendously.

			if eqloc == CsvEqSentinel {
				// Invalid field syntax: Drop the field but keep the record
				if verbose {
					Log.Infof(
						"Dropping field with bad form: %s",
						"(elided)", /*tokenizer.BufSubstringSlow(start, lim), - see NOTE above*/
					)
				}
				softErrors++
				continue FieldLoop
			}

			// No need to check that BufAt(start+1) is valid: The first two characters will
			// always be present because eqloc is either CsvEqSentinel (handled above) or
			// greater than zero (the field name is never empty).
			switch tokenizer.BufAt(start) {
			case 'A':
				switch tokenizer.BufAt(start + 1) {
				case 'c':
					if val, ok := match(tokenizer, start, lim, eqloc, "Account"); ok {
						info.Account = ustrs.AllocBytes(val)
						matched = true
					}
				case 'l':
					if val, ok := match(tokenizer, start, lim, eqloc, "AllocTRES"); ok {
						info.ReqGPUS, gputmp = ParseAllocTRES(val, ustrs, gputmp)
						matched = true
					}
				case 'v':
					if val, ok := match(tokenizer, start, lim, eqloc, "AveCPU"); ok {
						info.AveCPU, err = parseSlurmElapsed64(val)
						matched = true
					} else if val, ok := match(tokenizer, start, lim, eqloc, "AveDiskRead"); ok {
						info.AveDiskRead, err = parseSlurmBytes(val)
						matched = true
					} else if val, ok := match(tokenizer, start, lim, eqloc, "AveDiskWrite"); ok {
						info.AveDiskWrite, err = parseSlurmBytes(val)
						matched = true
					} else if val, ok := match(tokenizer, start, lim, eqloc, "AveRSS"); ok {
						info.AveRSS, err = parseSlurmBytes(val)
						matched = true
					} else if val, ok := match(tokenizer, start, lim, eqloc, "AveVMSize"); ok {
						info.AveVMSize, err = parseSlurmBytes(val)
						matched = true
					}
				}
			case 'E':
				if val, ok := match(tokenizer, start, lim, eqloc, "ElapsedRaw"); ok {
					info.ElapsedRaw, err = parseUint32(val)
					matched = true
				} else if val, ok := match(tokenizer, start, lim, eqloc, "ExitCode"); ok {
					sep := bytes.IndexByte(val, ':')
					if sep == -1 {
						// oh well
						info.ExitCode, err = parseUint8(val)
					} else {
						var e1, e2 error
						info.ExitCode, e1 = parseUint8(val[:sep])
						info.ExitSignal, e2 = parseUint8(val[sep+1:])
						err = errors.Join(e1, e2)
					}
					matched = true
				} else if val, ok := match(tokenizer, start, lim, eqloc, "End"); ok {
					info.End, err = parseRFC3339(val)
					matched = true
				}
			case 'J':
				if val, ok := match(tokenizer, start, lim, eqloc, "JobID"); ok {
					// If the ID indicates an array job then set ArrayJobID, ArrayIndex, and
					// ArrayStep.
					//
					// If the ID indicates a het job then set HetJobID, HetJobOffset, and
					// HetJobStep.
					//
					// In either case, JobID and JobStep are left alone - they are set from
					// JobIDRaw.

					sep := bytes.IndexByte(val, '.')
					step := UstrEmpty
					if sep != -1 {
						step = ustrs.AllocBytes(val[sep+1:])
						val = val[:sep]
					}
					specialSep := bytes.IndexAny(val, "+_")
					if specialSep != -1 {
						jobid, e1 := parseUint32(val[:specialSep])
						jobix, e2 := parseUint32(val[specialSep+1:])
						err = errors.Join(e1, e2)
						if val[specialSep] == '_' {
							info.ArrayJobID = jobid
							info.ArrayIndex = jobix
							info.ArrayStep = step
						} else {
							info.HetJobID = jobid
							info.HetOffset = jobix
							info.HetStep = step
						}
					}
					matched = true
				} else if val, ok := match(tokenizer, start, lim, eqloc, "JobIDRaw"); ok {
					sep := bytes.IndexByte(val, '.')
					if sep == -1 {
						info.JobID, err = parseUint32(val)
					} else {
						info.JobID, err = parseUint32(val[:sep])
						info.JobStep = ustrs.AllocBytes(val[sep+1:])
					}
					matched = true
				} else if val, ok := match(tokenizer, start, lim, eqloc, "JobName"); ok {
					info.JobName = ustrs.AllocBytes(val)
					matched = true
				}
			case 'L':
				if val, ok := match(tokenizer, start, lim, eqloc, "Layout"); ok {
					info.Layout = ustrs.AllocBytes(val)
					matched = true
				}
			case 'M':
				if val, ok := match(tokenizer, start, lim, eqloc, "MaxRSS"); ok {
					info.MaxRSS, err = parseSlurmBytes(val)
					matched = true
				} else if val, ok := match(tokenizer, start, lim, eqloc, "MaxVMSize"); ok {
					info.MaxVMSize, err = parseSlurmBytes(val)
					matched = true
				} else if val, ok := match(tokenizer, start, lim, eqloc, "MinCPU"); ok {
					info.MinCPU, err = parseSlurmElapsed64(val)
					matched = true
				}
			case 'N':
				if val, ok := match(tokenizer, start, lim, eqloc, "NodeList"); ok {
					info.NodeList = ustrs.AllocBytes(val)
					matched = true
				}
			case 'P':
				if val, ok := match(tokenizer, start, lim, eqloc, "Partition"); ok {
					info.Partition = ustrs.AllocBytes(val)
					matched = true
				} else if _, ok := match(tokenizer, start, lim, eqloc, "Priority"); ok {
					// No field for this yet
					matched = true
				}
			case 'R':
				if val, ok := match(tokenizer, start, lim, eqloc, "ReqCPUS"); ok {
					// Stick to the Slurm spelling: "ReqCPUS" with a capital S
					info.ReqCPUS, err = parseUint32(val)
					matched = true
				} else if val, ok := match(tokenizer, start, lim, eqloc, "ReqMem"); ok {
					info.ReqMem, err = parseSlurmBytes(val)
					matched = true
				} else if val, ok := match(tokenizer, start, lim, eqloc, "ReqNodes"); ok {
					info.ReqNodes, err = parseUint32(val)
					matched = true
				} else if val, ok := match(tokenizer, start, lim, eqloc, "Reservation"); ok {
					info.Reservation = ustrs.AllocBytes(val)
					matched = true
				}
			case 'S':
				if val, ok := match(tokenizer, start, lim, eqloc, "Start"); ok {
					info.Start, err = parseRFC3339(val)
					matched = true
				} else if val, ok := match(tokenizer, start, lim, eqloc, "State"); ok {
					// Strip data following the first word: "CANCELLED by x" becomes "CANCELLED".
					if loc := bytes.IndexByte(val, ' '); loc != -1 {
						val = val[:loc]
					}
					info.State = ustrs.AllocBytes(val)
					matched = true
				} else if val, ok := match(tokenizer, start, lim, eqloc, "Submit"); ok {
					info.Submit, err = parseRFC3339(val)
					matched = true
				} else if val, ok := match(tokenizer, start, lim, eqloc, "Suspended"); ok {
					info.Suspended, err = parseSlurmElapsed32(val)
					matched = true
				} else if val, ok := match(tokenizer, start, lim, eqloc, "SystemCPU"); ok {
					info.SystemCPU, err = parseSlurmElapsed64(val)
					matched = true
				}
			case 'T':
				if val, ok := match(tokenizer, start, lim, eqloc, "TimelimitRaw"); ok {
					info.TimelimitRaw, err = parseUint32(val) // input value is in minutes
					info.TimelimitRaw *= 60
					matched = true
				}
			case 'U':
				if val, ok := match(tokenizer, start, lim, eqloc, "User"); ok {
					info.User = ustrs.AllocBytes(val)
					matched = true
				} else if val, ok := match(tokenizer, start, lim, eqloc, "UserCPU"); ok {
					info.UserCPU, err = parseSlurmElapsed64(val)
					matched = true
				}
			case 'v':
				if val, ok := match(tokenizer, start, lim, eqloc, "v"); ok {
					info.Version = ustrs.AllocBytes(val)
					matched = true
				}
			}

			// Four cases:
			//
			//   matched && !failed - field matched a tag, value is good
			//   matched && failed - field matched a tag, value is bad
			//   !matched && !failed - field did not match any tag
			//   !matched && failed - impossible
			//
			// The second case suggests something bad, so discard the record in this case.  Note
			// this is actually the same as just `failed` due to the fourth case.
			if matched {
				anyMatched = true
			} else {
				if verbose {
					Log.Warningf(
						"Dropping field with unknown name: %s",
						"(elided)", /* tokenizer.BufSubstringSlow(start, eqloc-1), -
						   see NOTE above */
					)
				}
				if err == nil {
					softErrors++
				}
			}
			if err != nil {
				if verbose {
					Log.Warningf(
						"Dropping record with illegal/unparseable value: %s %v",
						"(elided)", /*tokenizer.BufSubstringSlow(start, lim), - see NOTE above */
						err,
					)
				}
				softErrors++
				tokenizer.ScanEol()
				continue LineLoop
			}
		} // end FieldLoop

		if !anyMatched && endOfInput {
			continue LineLoop
		}

		// Fields have been parsed, now check them

		irritants := ""
		if info.Version == UstrEmpty || info.End == math.MaxInt64 || info.JobID == 0 {
			if info.Version == UstrEmpty {
				irritants += "version "
			}
			if info.End == math.MaxInt64 {
				irritants += "end "
			}
			if info.JobID == 0 {
				irritants += "jobid "
			}
		}
		if irritants != "" {
			if verbose {
				Log.Warningf("Dropping record with missing mandatory field(s): %s", irritants)
			}
			softErrors++
			continue LineLoop
		}

		// Keep it

		records = append(records, info)
	}

	err = nil
	return
}
