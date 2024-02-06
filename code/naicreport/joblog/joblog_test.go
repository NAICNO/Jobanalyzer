package joblog

import (
	"fmt"
	"os"
	"testing"
	"time"

	"go-utils/filesys"
	"go-utils/freecsv"
	sonartime "go-utils/time"
)

func TestReadLogFiles(t *testing.T) {
	dataPath, err := filesys.PopulateTestData(
		"joblog",
		filesys.TestFile{"2023/09/03", "jobdata.csv", []byte(file_2023_09_03)},
		filesys.TestFile{"2023/09/06", "jobdata.csv", []byte(file_2023_09_06)},
		filesys.TestFile{"2023/09/07", "jobdata.csv", []byte(file_2023_09_07)},
	)
	defer os.RemoveAll(dataPath)
	if err != nil {
		t.Fatalf("Could not create test hierarchy: %q", err)
	}

	// The file on September 3 has only one record, so we should find that

	from := time.Date(2023, 9, 3, 0, 0, 0, 0, time.UTC)
	to := time.Date(2023, 9, 4, 0, 0, 0, 0, time.UTC)
	files, err := FindJoblogFiles(dataPath, "jobdata.csv", from, to)
	if err != nil {
		t.Fatalf("Could not enumerate files: %q", err)
	}
	jobLog, err := ReadJoblogFiles[*testJob](
		files,
		false,
		parseJobRecord,
		integrateJobRecords,
	)
	if err != nil {
		t.Fatalf("Could not read: %q", err)
	}

	if len(jobLog) != 1 {
		t.Fatalf("Unexpected job log length %d", len(jobLog))
	}
	xs := findJobsByHostAndId(jobLog, "ml6", 2166356)
	if len(xs) != 1 {
		t.Fatalf("Did not find exactly one record for (%s,%d): %d\n%v", "ml6", 2166356, len(xs), xs)
	}
	x := xs[0]
	if x.Id != 2166356 || x.Host != "ml6" || x.user != "user3" ||
		x.firstSeen != time.Date(2023, 9, 3, 20, 0, 0, 0, time.UTC) ||
		x.LastSeen != time.Date(2023, 9, 3, 20, 0, 0, 0, time.UTC) ||
		x.start != time.Date(2023, 9, 3, 15, 10, 0, 0, time.UTC) ||
		x.end != time.Date(2023, 9, 3, 16, 50, 0, 0, time.UTC) {
		fmt.Fprintf(os.Stderr, "first=%v last=%v start=%v end=%v\n", x.firstSeen, x.LastSeen, x.start, x.end)
		t.Fatalf("#1: Bad record %v", x)
	}

	// The files on September 6 and 7 have jobs spanning the two days.

	from = time.Date(2023, 9, 6, 0, 0, 0, 0, time.UTC)
	to = time.Date(2023, 9, 8, 0, 0, 0, 0, time.UTC)
	files, err = FindJoblogFiles(dataPath, "jobdata.csv", from, to)
	if err != nil {
		t.Fatalf("Could not enumerate files: %q", err)
	}
	jobLog, err = ReadJoblogFiles(
		files,
		false,
		parseJobRecord,
		integrateJobRecords,
	)
	if err != nil {
		t.Fatalf("Could not read: %q", err)
	}

	// For Job 2253420 there should just be a single record.

	xs = findJobsByHostAndId(jobLog, "ml8", 2253420)
	if len(xs) != 1 {
		t.Fatalf("Did not find exactly one record for (%s,%d): %d", "ml8", 2253420, len(xs))
	}
	x = xs[0]
	if x.Id != 2253420 || x.Host != "ml8" || x.user != "user1" ||
		x.firstSeen != time.Date(2023, 9, 5, 22, 0, 0, 0, time.UTC) ||
		x.LastSeen != time.Date(2023, 9, 7, 14, 0, 0, 0, time.UTC) ||
		x.start != time.Date(2023, 9, 5, 16, 5, 0, 0, time.UTC) ||
		x.end != time.Date(2023, 9, 6, 16, 30, 0, 0, time.UTC) {
		fmt.Fprintf(os.Stderr, "first=%v last=%v start=%v end=%v\n", x.firstSeen, x.LastSeen, x.start, x.end)
		t.Fatalf("#2: Bad record %v", x)
	}

	// For Job 2712710 there is a gap in the timeline on sept 7 at 1200 so there should be two records.

	xs = findJobsByHostAndId(jobLog, "ml6", 2712710)
	if len(xs) != 2 {
		t.Fatalf("Did not find exactly two records for (%s,%d): %d", "ml6", 2712710, len(xs))
	}
	x0 := xs[0]
	x1 := xs[1]

	// Fairly basic test: x0 should be the younger one and x1 the older one

	if !x0.LastSeen.After(x1.LastSeen) {
		t.Fatalf("Records in wrong order")
	}
}

func findJobsByHostAndId(jobLog []*JobsByHost[*testJob], host string, id uint32) []*testJob {
	jobs := []*testJob{}
	for _, x := range jobLog {
		if x.Host == host {
			for _, y := range x.Jobs {
				if y.Id == id {
					jobs = append(jobs, y)
				}
			}
		}
	}
	return jobs
}

// Structure of the test jobs

type testJob struct {
	GenericJob
	user      string    // user's login name
	firstSeen time.Time // timestamp of record in which job is first seen
	start     time.Time // the start field of the first record for the job
	end       time.Time // the end field of the last record for the job
}

func parseJobRecord(r map[string]string) (*testJob, bool) {
	success := true

	tag := freecsv.GetString(r, "tag", &success)
	success = success && tag == "mytag"
	timestamp := freecsv.GetSonarDateTime(r, "now", &success)
	id := freecsv.GetJobMark(r, "jobm", &success)
	user := freecsv.GetString(r, "user", &success)
	host := freecsv.GetString(r, "host", &success)
	start := freecsv.GetSonarDateTime(r, "start", &success)
	end := freecsv.GetSonarDateTime(r, "end", &success)

	if !success {
		return nil, false
	}

	return &testJob{
		GenericJob: GenericJob{
			Id:       id,
			Host:     host,
			LastSeen: timestamp,
			Expired:  false,
		},
		user:      user,
		firstSeen: timestamp,
		start:     start,
		end:       end,
	}, true
}

func integrateJobRecords(record, probe *testJob) {
	record.LastSeen = sonartime.MaxTime(record.LastSeen, probe.LastSeen)
	record.firstSeen = sonartime.MinTime(record.firstSeen, probe.firstSeen)
	record.start = sonartime.MinTime(record.start, probe.start)
	record.end = sonartime.MaxTime(record.end, probe.end)
}

// These are real cpuhog data, but anonymized and slightly cleaned up to deal with some artifacts
// from early runs.  A bunch of fields are not used by the test case but it didn't seem necessary to
// remove them.

const (
	file_2023_09_03 = `now=2023-09-03 20:00,jobm=2166356,user=user3,duration=0d 1h40m,host=ml6,cpu-peak=2615,gpu-peak=0,rcpu-avg=3,rcpu-peak=41,rmem-avg=12,rmem-peak=14,start=2023-09-03 15:10,end=2023-09-03 16:50,cmd=python3.9,tag=mytag
`

	file_2023_09_06 = `now=2023-09-05 22:00,jobm=2253420>,user=user1,duration=0d 5h50m,host=ml8,cpu-peak=13877,gpu-peak=0,rcpu-avg=29,rcpu-peak=73,rmem-avg=53,rmem-peak=54,start=2023-09-05 16:05,end=2023-09-05 21:55,cmd=python3,tag=mytag
now=2023-09-06 00:00,jobm=2253420,user=user1,duration=0d 7h50m,host=ml8,cpu-peak=14421,gpu-peak=0,rcpu-avg=40,rcpu-peak=76,rmem-avg=53,rmem-peak=54,start=2023-09-05 16:05,end=2023-09-05 23:55,cmd=python3,tag=mytag
now=2023-09-06 02:00,jobm=2253420>,user=user1,duration=0d 9h50m,host=ml8,cpu-peak=14421,gpu-peak=0,rcpu-avg=41,rcpu-peak=76,rmem-avg=53,rmem-peak=54,start=2023-09-05 16:05,end=2023-09-06 01:55,cmd=python3,tag=mytag
now=2023-09-06 04:00,jobm=2253420>,user=user1,duration=0d11h50m,host=ml8,cpu-peak=14421,gpu-peak=0,rcpu-avg=40,rcpu-peak=76,rmem-avg=53,rmem-peak=54,start=2023-09-05 16:05,end=2023-09-06 03:55,cmd=python3,tag=mytag
now=2023-09-06 06:00,jobm=2253420>,user=user1,duration=0d13h50m,host=ml8,cpu-peak=14421,gpu-peak=0,rcpu-avg=43,rcpu-peak=76,rmem-avg=53,rmem-peak=54,start=2023-09-05 16:05,end=2023-09-06 05:55,cmd=python3,tag=mytag
now=2023-09-06 08:00,jobm=2253420,user=user1,duration=0d15h50m,host=ml8,cpu-peak=14421,gpu-peak=0,rcpu-avg=42,rcpu-peak=76,rmem-avg=53,rmem-peak=54,start=2023-09-05 16:05,end=2023-09-06 07:55,cmd=python3,tag=mytag
now=2023-09-06 10:00,jobm=2253420,user=user1,duration=0d17h50m,host=ml8,cpu-peak=14421,gpu-peak=0,rcpu-avg=40,rcpu-peak=76,rmem-avg=53,rmem-peak=54,start=2023-09-05 16:05,end=2023-09-06 09:55,cmd=python3,tag=mytag
now=2023-09-06 12:00,jobm=2253420,user=user1,duration=0d19h50m,host=ml8,cpu-peak=14421,gpu-peak=0,rcpu-avg=43,rcpu-peak=76,rmem-avg=53,rmem-peak=54,start=2023-09-05 16:05,end=2023-09-06 11:55,cmd=python3,tag=mytag
now=2023-09-06 12:00,jobm=2712710,user=user2,duration=0d 4h20m,host=ml6,cpu-peak=1274,gpu-peak=0,rcpu-avg=3,rcpu-peak=20,rmem-avg=1,rmem-peak=2,start=2023-09-06 07:35,end=2023-09-06 11:55,cmd=kited,tag=mytag
now=2023-09-06 14:00,jobm=2253420,user=user1,duration=0d21h50m,host=ml8,cpu-peak=14421,gpu-peak=0,rcpu-avg=42,rcpu-peak=76,rmem-avg=54,rmem-peak=54,start=2023-09-05 16:05,end=2023-09-06 13:55,cmd=python3,tag=mytag
now=2023-09-06 14:00,jobm=2712710,user=user2,duration=0d 6h20m,host=ml6,cpu-peak=1274,gpu-peak=0,rcpu-avg=2,rcpu-peak=20,rmem-avg=1,rmem-peak=2,start=2023-09-06 07:35,end=2023-09-06 13:55,cmd=kited,tag=mytag
now=2023-09-06 16:00,jobm=2253420>,user=user1,duration=0d23h50m,host=ml8,cpu-peak=14421,gpu-peak=0,rcpu-avg=40,rcpu-peak=76,rmem-avg=54,rmem-peak=54,start=2023-09-05 16:05,end=2023-09-06 15:55,cmd=python3,tag=mytag
now=2023-09-06 16:00,jobm=2712710>,user=user2,duration=0d 8h20m,host=ml6,cpu-peak=1274,gpu-peak=0,rcpu-avg=2,rcpu-peak=20,rmem-avg=2,rmem-peak=2,start=2023-09-06 07:35,end=2023-09-06 15:55,cmd=kited,tag=mytag
now=2023-09-06 18:00,jobm=2253420,user=user1,duration=0d22h30m,host=ml8,cpu-peak=14421,gpu-peak=0,rcpu-avg=40,rcpu-peak=76,rmem-avg=54,rmem-peak=54,start=2023-09-05 18:00,end=2023-09-06 16:30,cmd=python3,tag=mytag
now=2023-09-06 18:00,jobm=2712710,user=user2,duration=0d10h20m,host=ml6,cpu-peak=1274,gpu-peak=0,rcpu-avg=2,rcpu-peak=20,rmem-avg=2,rmem-peak=2,start=2023-09-06 07:35,end=2023-09-06 17:55,cmd=kited,tag=mytag
now=2023-09-06 20:00,jobm=2253420,user=user1,duration=0d20h25m,host=ml8,cpu-peak=14421,gpu-peak=0,rcpu-avg=42,rcpu-peak=76,rmem-avg=54,rmem-peak=54,start=2023-09-05 20:05,end=2023-09-06 16:30,cmd=python3,tag=mytag
now=2023-09-06 20:00,jobm=2712710>,user=user2,duration=0d12h25m,host=ml6,cpu-peak=1274,gpu-peak=0,rcpu-avg=1,rcpu-peak=20,rmem-avg=2,rmem-peak=2,start=2023-09-06 07:35,end=2023-09-06 20:00,cmd=kited,tag=mytag
`

	file_2023_09_07 = `now=2023-09-06 22:00,jobm=2253420,user=user1,duration=0d18h30m,host=ml8,cpu-peak=14421,gpu-peak=0,rcpu-avg=42,rcpu-peak=76,rmem-avg=54,rmem-peak=54,start=2023-09-05 22:00,end=2023-09-06 16:30,cmd=python3,tag=mytag
now=2023-09-06 22:00,jobm=2712710>,user=user2,duration=0d14h20m,host=ml6,cpu-peak=1274,gpu-peak=0,rcpu-avg=1,rcpu-peak=20,rmem-avg=2,rmem-peak=2,start=2023-09-06 07:35,end=2023-09-06 21:55,cmd=kited,tag=mytag
now=2023-09-07 00:00,jobm=2253420,user=user1,duration=0d16h30m,host=ml8,cpu-peak=14240,gpu-peak=0,rcpu-avg=39,rcpu-peak=75,rmem-avg=54,rmem-peak=54,start=2023-09-06 00:00,end=2023-09-06 16:30,cmd=python3,tag=mytag
now=2023-09-07 00:00,jobm=2712710,user=user2,duration=0d16h20m,host=ml6,cpu-peak=1274,gpu-peak=0,rcpu-avg=1,rcpu-peak=20,rmem-avg=2,rmem-peak=2,start=2023-09-06 07:35,end=2023-09-06 23:55,cmd=kited,tag=mytag
now=2023-09-07 02:00,jobm=2253420,user=user1,duration=0d14h30m,host=ml8,cpu-peak=14240,gpu-peak=0,rcpu-avg=38,rcpu-peak=75,rmem-avg=54,rmem-peak=54,start=2023-09-06 02:00,end=2023-09-06 16:30,cmd=python3,tag=mytag
now=2023-09-07 02:00,jobm=2712710,user=user2,duration=0d18h20m,host=ml6,cpu-peak=1274,gpu-peak=0,rcpu-avg=1,rcpu-peak=20,rmem-avg=2,rmem-peak=2,start=2023-09-06 07:35,end=2023-09-07 01:55,cmd=kited,tag=mytag
now=2023-09-07 04:00,jobm=2253420,user=user1,duration=0d12h30m,host=ml8,cpu-peak=14240,gpu-peak=0,rcpu-avg=38,rcpu-peak=75,rmem-avg=54,rmem-peak=54,start=2023-09-06 04:00,end=2023-09-06 16:30,cmd=python3,tag=mytag
now=2023-09-07 04:00,jobm=2712710,user=user2,duration=0d20h20m,host=ml6,cpu-peak=1274,gpu-peak=0,rcpu-avg=1,rcpu-peak=20,rmem-avg=2,rmem-peak=2,start=2023-09-06 07:35,end=2023-09-07 03:55,cmd=kited,tag=mytag
now=2023-09-07 06:00,jobm=2253420,user=user1,duration=0d10h30m,host=ml8,cpu-peak=14240,gpu-peak=0,rcpu-avg=34,rcpu-peak=75,rmem-avg=54,rmem-peak=54,start=2023-09-06 06:00,end=2023-09-06 16:30,cmd=python3,tag=mytag
now=2023-09-07 06:00,jobm=2712710>,user=user2,duration=0d22h20m,host=ml6,cpu-peak=1274,gpu-peak=0,rcpu-avg=1,rcpu-peak=20,rmem-avg=2,rmem-peak=2,start=2023-09-06 07:35,end=2023-09-07 05:55,cmd=kited,tag=mytag
now=2023-09-07 08:00,jobm=2253420,user=user1,duration=0d 8h30m,host=ml8,cpu-peak=14240,gpu-peak=0,rcpu-avg=33,rcpu-peak=75,rmem-avg=54,rmem-peak=54,start=2023-09-06 08:00,end=2023-09-06 16:30,cmd=python3,tag=mytag
now=2023-09-07 08:00,jobm=2712710,user=user2,duration=0d23h55m,host=ml6,cpu-peak=1274,gpu-peak=0,rcpu-avg=1,rcpu-peak=20,rmem-avg=2,rmem-peak=2,start=2023-09-06 08:05,end=2023-09-07 08:00,cmd=kited,tag=mytag
now=2023-09-07 08:00,jobm=3043187,user=user3,duration=0d 0h10m,host=ml6,cpu-peak=1519,gpu-peak=0,rcpu-avg=8,rcpu-peak=24,rmem-avg=3,rmem-peak=3,start=2023-09-07 07:50,end=2023-09-07 08:00,cmd=python3.9,tag=mytag
now=2023-09-07 10:00,jobm=2712710,user=user2,duration=0d23h55m,host=ml6,cpu-peak=1274,gpu-peak=0,rcpu-avg=1,rcpu-peak=20,rmem-avg=2,rmem-peak=2,start=2023-09-06 10:00,end=2023-09-07 09:55,cmd=kited,tag=mytag
now=2023-09-07 10:00,jobm=2253420,user=user1,duration=0d 6h30m,host=ml8,cpu-peak=14240,gpu-peak=0,rcpu-avg=36,rcpu-peak=75,rmem-avg=53,rmem-peak=54,start=2023-09-06 10:00,end=2023-09-06 16:30,cmd=python3,tag=mytag
now=2023-09-07 10:00,jobm=3043187,user=user3,duration=0d 0h20m,host=ml6,cpu-peak=1519,gpu-peak=0,rcpu-avg=5,rcpu-peak=24,rmem-avg=3,rmem-peak=3,start=2023-09-07 07:50,end=2023-09-07 08:10,cmd=python3.9,tag=mytag
now=2023-09-07 12:00,jobm=2253420,user=user1,duration=0d 4h30m,host=ml8,cpu-peak=13528,gpu-peak=0,rcpu-avg=22,rcpu-peak=71,rmem-avg=53,rmem-peak=54,start=2023-09-06 12:00,end=2023-09-06 16:30,cmd=python3,tag=mytag
now=2023-09-07 12:00,jobm=3043187,user=user3,duration=0d 0h20m,host=ml6,cpu-peak=1519,gpu-peak=0,rcpu-avg=5,rcpu-peak=24,rmem-avg=3,rmem-peak=3,start=2023-09-07 07:50,end=2023-09-07 08:10,cmd=python3.9,tag=mytag
now=2023-09-07 14:00,jobm=2253420,user=user1,duration=0d 2h30m,host=ml8,cpu-peak=8423,gpu-peak=0,rcpu-avg=16,rcpu-peak=44,rmem-avg=53,rmem-peak=54,start=2023-09-06 14:00,end=2023-09-06 16:30,cmd=python3,tag=mytag
now=2023-09-07 14:00,jobm=2712710,user=user2,duration=0d23h55m,host=ml6,cpu-peak=761,gpu-peak=0,rcpu-avg=1,rcpu-peak=12,rmem-avg=2,rmem-peak=2,start=2023-09-06 14:00,end=2023-09-07 13:55,cmd=kited,tag=mytag
now=2023-09-07 14:00,jobm=3043187,user=user3,duration=0d 0h20m,host=ml6,cpu-peak=1519,gpu-peak=0,rcpu-avg=5,rcpu-peak=24,rmem-avg=3,rmem-peak=3,start=2023-09-07 07:50,end=2023-09-07 08:10,cmd=python3.9,tag=mytag
now=2023-09-07 14:00,jobm=3129396,user=user3,duration=0d 0h55m,host=ml6,cpu-peak=3756,gpu-peak=0,rcpu-avg=13,rcpu-peak=59,rmem-avg=3,rmem-peak=3,start=2023-09-07 13:00,end=2023-09-07 13:55,cmd=python3.9,tag=mytag
now=2023-09-07 21:27,jobm=3043187,user=user3,duration=0d 0h20m,host=ml6,cpu-peak=1519,gpu-peak=0,rcpu-avg=5,rcpu-peak=24,rmem-avg=3,rmem-peak=3,start=2023-09-07 07:50,end=2023-09-07 08:10,cmd=python3.9,tag=mytag
`
)
