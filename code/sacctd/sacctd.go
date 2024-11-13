// Run sacct with sensible parameters, select data, and dump output on stdout on free csv format,
// suitable for transmission to jobanalyzer.
//
// Usage:
//  sacctd [-window n] [-span n,m] [input-file]
//
// where
//  -window minutes
//    Set the start time to now-minutes and the end time to now, and dump records that are relevant
//    for that interval.  Normally the running interval of sacctd should be less than the window, so
//    there will be some overhead in what's reported; the analysis host will have to remove
//    redundancies.  The default window is configurable below.  Precludes -span.
//
//  -span yyyy-mm-dd,yyyy-mm-dd
//    From-to dates; `from` is inclusive, `to` is exclusive.  Precludes -window
//
//  input-file
//    For testing: If present, read input from this instead of running sacct.  Ignores -window
//    and -span (beyond requiring them to be valid or absent).
//
// Notes for users:
//
// Fields that are empty, zero, or unknown are generally elided.
//
// All timestamps are reformatted to include time zone.
//
// The following triplet (reformatted) is instructive (missing newer "JobIDRaw", includes obsoleted
// "time"):
//
//   v=0.1.0,time=2024-08-05T10:00:44+02:00,JobID=756717,User=ec313-autotekst,Account=ec313,
//     State=COMPLETED,Start=2024-08-05T09:01:08+02:00,End=2024-08-05T09:01:39+02:00,
//     ElapsedRaw=31,ReqCPUS=8,ReqMem=64G,ReqNodes=1,Submit=2024-08-05T09:01:07+02:00,
//     SystemCPU=00:09.522,TimelimitRaw=2,UserCPU=00:17.122,NodeList=gpu-10,
//     JobName=vms-transcription-job
//   v=0.1.0,time=2024-08-05T10:00:44+02:00,JobID=756717.batch,Account=ec313,State=COMPLETED,
//     Start=2024-08-05T09:01:08+02:00,End=2024-08-05T09:01:39+02:00,AveCPU=00:00:26,
//     AveDiskRead=5098.29M,AveDiskWrite=3.69M,AveRSS=5135468K,ElapsedRaw=31,MaxRSS=5135468K,
//     MinCPU=00:00:26,ReqCPUS=8,ReqNodes=1,Submit=2024-08-05T09:01:08+02:00,SystemCPU=00:09.521,
//     UserCPU=00:17.122,NodeList=gpu-10,JobName=batch
//   v=0.1.0,time=2024-08-05T10:00:44+02:00,JobID=756717.extern,Account=ec313,State=COMPLETED,
//     Start=2024-08-05T09:01:08+02:00,End=2024-08-05T09:01:39+02:00,AveDiskRead=0.01M,
//     ElapsedRaw=31,ReqCPUS=8,ReqNodes=1,Submit=2024-08-05T09:01:08+02:00,SystemCPU=00:00.001,
//     NodeList=gpu-10,JobName=extern
//
// Various steps provide various data; the consumer must integrate these data somehow.
//
// Both JobID and JobIDRaw are printed as these reveal different types of information and it's
// not known at this time what we'll need.
//
// Re the batch and extern steps, there is a good explanation here:
//
//    https://stackoverflow.com/questions/52447602/slurm-sacct-shows-batch-and-extern-job-names
//
// (Note especially the comment "Many non-MPI programs do a lot of calculations in the batch step,
// so the resource usage is accounted there.")

package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const (
	version = "0.1.0"

	// 90 minutes is an OK window if we expect to be running this every hour.
	defaultWindow = 90
	defaultSpan   = ""
)

var (
	span   = flag.String("span", "", "[From-to) time span yyyy-mm-dd,yyyy-mm-dd")
	window = flag.Uint("window", defaultWindow, "Window (in minutes) for the sacct query")
)

var (
	// Examine terminated jobs only
	states = []string{
		"CANCELLED",
		"COMPLETED",
		"DEADLINE",
		"FAILED",
		"OUT_OF_MEMORY",
		"TIMEOUT",
	}

	// Basically, we can just pile it on here, but it's unlikely that everything is of interest,
	// hence we select.  The capitalization should be exactly as it is in the sacct man page, though
	// sacct appears to ignore capitalization.
	fieldNames = []string{
		"JobID",
		"JobIDRaw",
		"User",
		"Account",
		"State",
		"Start",
		"End",
		"AveCPU",
		"AveDiskRead",
		"AveDiskWrite",
		"AveRSS",
		"AveVMSize",
		"ElapsedRaw",
		"ExitCode",
		"Layout",
		"MaxRSS",
		"MaxVMSize",
		"MinCPU",
		"ReqCPUS",
		"ReqMem",
		"ReqNodes",
		"Reservation",
		"Submit",
		"Suspended",
		"SystemCPU",
		"TimelimitRaw",
		"UserCPU",
		"NodeList",
		"Partition",
		"AllocTRES",
		// JobName should always be last in case it contains `|`, code below will clean that up.
		"JobName",
	}

	// Fields that are dates that may be reinterpreted before transmission.
	isDateField = map[string]bool{
		"Start":  true,
		"End":    true,
		"Submit": true,
	}

	spanRe = regexp.MustCompile(`^(\d\d\d\d-\d\d-\d\d),(\d\d\d\d-\d\d-\d\d)$`)
)

func main() {
	flag.Parse()

	if *window != defaultWindow && *span != defaultSpan {
		log.Fatal("Can't use both -window and -span")
	}
	var from, to string
	if *span != defaultSpan {
		matches := spanRe.FindStringSubmatch(*span)
		if len(matches) != 3 {
			log.Fatalf("Invalid span %s", *span)
		}
		from = matches[1]
		to = matches[2]
	} else {
		from = fmt.Sprintf("now-%dminutes", *window)
		to = "now"
	}

	var sacct_output string
	if len(flag.Args()) > 0 {
		if len(flag.Args()) > 1 {
			log.Fatalf("At most one input file")
		}
		bytes, err := os.ReadFile(flag.Args()[0])
		if err != nil {
			log.Fatal(err)
		}
		sacct_output = string(bytes)
	} else {
		cmd := exec.Command(
			"sacct",
			"-aP",
			"-s",
			strings.Join(states, ","),
			"--noheader",
			"-o",
			strings.Join(fieldNames, ","),
			"-S",
			from,
			"-E",
			to,
		)
		var out strings.Builder
		cmd.Stdout = &out
		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
		sacct_output = out.String()
	}

	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	scan := bufio.NewScanner(strings.NewReader(sacct_output))
	versionField := "v=" + version
	for scan.Scan() {
		fields := strings.Split(scan.Text(), "|")

		// If there are more fields than field names then that's because the job name contains `|`.
		// The JobName field always comes last.  Catenate excess fields until we have the same
		// number of fields and names.  (Could just ignore excess fields instead.)
		for len(fields) > len(fieldNames) {
			fields[len(fields)-2] += "|" + fields[len(fields)-1]
			fields = fields[:len(fields)-1]
		}

		// Each record can be a different length b/c zero fields.  It's wastful to remake the array
		// every time around the loop but it doesn't matter.
		csvfields := make([]string, 0, len(fields)+1)
		csvfields = append(csvfields, versionField)
		for i, n := range fieldNames {
			val := fields[i]
			if !isZero(n, val) {
				if isDateField[n] {
					// The slurm date format is localtime without a time zone offset.  This is bound
					// to lead to problems eventually, so reformat.  If parsing fails, just transmit
					// the date and let the consumer deal with it.
					t, err := time.ParseInLocation("2006-01-02T15:04:05", val, time.Local)
					if err == nil {
						val = t.Format(time.RFC3339)
					}
				}
				csvfields = append(csvfields, n+"="+val)
			}
		}

		w.Write(csvfields)
	}
}

var (
	// Various zero values in "controlled" fields.
	zero = map[string]bool{
		"Unknown":  true,
		"0":        true,
		"00:00:00": true,
		"0:0":      true,
		"0.00M":    true,
	}

	// These fields may contain zero values that don't mean zero.
	uncontrolled = map[string]bool{
		"JobName": true,
		"Account": true,
		"User":    true,
	}
)

func isZero(field, value string) bool {
	return value == "" || (!uncontrolled[field] && zero[value])
}
