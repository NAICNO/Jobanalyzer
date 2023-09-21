package mlcpuhog

import (
	"os"
	"path"
	"testing"
	"time"

	"naicreport/jobstate"
)

func TestReadLogFiles(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %q", err)
	}

	// The file on September 3 has only one record
	dataPath := path.Join(wd, "../../sonar_test_data0")
	from := time.Date(2023, 9, 3, 0, 0, 0, 0, time.UTC)
	to := time.Date(2023, 9, 4, 0, 0, 0, 0, time.UTC)
	jobLog, err := readCpuhogLogFiles(dataPath, from, to, false)
	if err != nil {
		t.Fatalf("Could not read: %q", err)
	}

	if len(jobLog) != 1 {
		t.Fatalf("Unexpected job log length %d", len(jobLog))
	}
	x, found := jobLog[jobstate.JobKey{Id: 2166356, Host: "ml6"}]
	if !found {
		t.Fatalf("Could not find record")
	}
	if x.id != 2166356 || x.host != "ml6" || x.user != "poyenyt" || x.cmd != "python3.9" ||
		x.firstSeen != time.Date(2023, 9, 3, 20, 0, 0, 0, time.UTC) ||
		x.lastSeen != time.Date(2023, 9, 3, 20, 0, 0, 0, time.UTC) ||
		x.start != time.Date(2023, 9, 3, 15, 10, 0, 0, time.UTC) ||
		x.end != time.Date(2023, 9, 3, 16, 50, 0, 0, time.UTC) ||
		x.cpuPeak != 2615 || x.gpuPeak != 0 || x.rcpuAvg != 3 || x.rcpuPeak != 41 ||
		x.rmemAvg != 12 || x.rmemPeak != 14 {
		t.Fatalf("Bad record %v", x)
	}

	// Then redo the operation but now pick up two files, one has a newer record for the job.
	// We want to see some of the old data but also some of the new data.

	// The files on September 6 and 7 have a job spanning the two.  (The job is not done at that
	// point but we should not see later data.)

	from = time.Date(2023, 9, 6, 0, 0, 0, 0, time.UTC)
	to = time.Date(2023, 9, 8, 0, 0, 0, 0, time.UTC)
	jobLog, err = readCpuhogLogFiles(dataPath, from, to, false)
	if err != nil {
		t.Fatalf("Could not read: %q", err)
	}

	x, found = jobLog[jobstate.JobKey{Id: 2712710, Host: "ml6"}]
	if !found {
		t.Fatalf("Could not find record")
	}

	if x.id != 2712710 || x.host != "ml6" || x.user != "hermanno" || x.cmd != "kited" ||
		x.firstSeen != time.Date(2023, 9, 6, 12, 0, 0, 0, time.UTC) ||
		x.lastSeen != time.Date(2023, 9, 7, 14, 0, 0, 0, time.UTC) ||
		x.start != time.Date(2023, 9, 6, 7, 35, 0, 0, time.UTC) ||
		x.end != time.Date(2023, 9, 7, 13, 55, 0, 0, time.UTC) ||
		x.cpuPeak != 1274 || x.gpuPeak != 0 || x.rcpuAvg != 3 || x.rcpuPeak != 20 ||
		x.rmemAvg != 2 || x.rmemPeak != 2 {
		t.Fatalf("Bad record %v", x)
	}

}
