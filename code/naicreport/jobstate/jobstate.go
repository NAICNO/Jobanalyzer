// Data persistence for various subsystems that track job state.
//
// The job information is represented on disk in free CSV form.  This means there's some annoying
// serialization and deserialization work, but the data are textual and structured at the same time,
// and this is better for debugging, resilience, and growth, at least for now.  In the future, maybe
// we'll use a gob instead, or a proper database.

package jobstate

import (
	"os"
	"strconv"
	"time"

	"go-utils/freecsv"
)

// Information about CPU hogs stored in the persistent state.  Other data that are needed for
// generating the report can be picked up from the log data for the job ID.

type JobDatabase struct {
	Active  map[JobKey]*JobState
	Expired map[ExpiredJobKey]*JobState
}

func NewJobDatabase() *JobDatabase {
	return &JobDatabase{
		Active:  make(map[JobKey]*JobState),
		Expired: make(map[ExpiredJobKey]*JobState),
	}
}

type JobState struct {
	Id                uint32
	Host              string
	StartedOnOrBefore time.Time
	FirstViolation    time.Time
	LastSeen          time.Time
	IsReported        bool
	Aux               any // in-memory value
}

// On the ML nodes, (job#, host) identifies a job uniquely because job#s are not coordinated across
// hosts and no job is cross-host.

type JobKey struct {
	Id   uint32
	Host string
}

type ExpiredJobKey struct {
	Id       uint32
	Host     string
	LastSeen time.Time
}

// Read the job state from disk and return a parsed and error-checked data structure.  Bogus records
// are silently dropped.
//
// If this returns an error, it is the error returned from freecsv.ReadFreeCSV, see that for more
// information.  No new errors are generated here.

func ReadJobDatabase(stateFilename string) (*JobDatabase, error, int) {
	stateCsv, err := freecsv.ReadFreeCSV(stateFilename)
	if err != nil {
		return nil, err, 0
	}
	db := NewJobDatabase()
	dropped := 0
	for _, repr := range stateCsv {
		success := true
		id := freecsv.GetUint32(repr, "id", &success)
		host := freecsv.GetString(repr, "host", &success)
		startedOnOrBefore := freecsv.GetRFC3339(repr, "startedOnOrBefore", &success)
		firstViolation := freecsv.GetRFC3339(repr, "firstViolation", &success)
		lastSeen := freecsv.GetRFC3339(repr, "lastSeen", &success)
		isReported := freecsv.GetBool(repr, "isReported", &success)
		isExpired := freecsv.GetBool(repr, "isExpired", new(bool))

		if !success {
			// Bogus record
			dropped++
			continue
		}
		job := &JobState{
			Id:                id,
			Host:              host,
			StartedOnOrBefore: startedOnOrBefore,
			FirstViolation:    firstViolation,
			LastSeen:          lastSeen,
			IsReported:        isReported,
		}
		if isExpired {
			db.Expired[ExpiredJobKey{id, host, lastSeen}] = job
		} else {
			db.Active[JobKey{id, host}] = job
		}
	}
	return db, nil, dropped
}

func ReadJobDatabaseOrEmpty(stateFilename string) (*JobDatabase, error, int) {
	db, err, dropped := ReadJobDatabase(stateFilename)
	if err == nil {
		return db, nil, 0
	}
	_, isPathErr := err.(*os.PathError)
	if isPathErr {
		return NewJobDatabase(), nil, 0
	}
	return nil, err, dropped
}

// If state does not have the job then add it.  In either case set its LastSeen field to lastSeen.
// Return true if added, false if not.

func EnsureJob(db *JobDatabase, id uint32, host string,
	started, firstViolation, lastSeen time.Time, expired bool, aux any) bool {
	job := &JobState{
		Id:                id,
		Host:              host,
		StartedOnOrBefore: started,
		FirstViolation:    firstViolation,
		LastSeen:          lastSeen,
		IsReported:        false,
		Aux:               aux,
	}
	if expired {
		k := ExpiredJobKey{Id: id, Host: host, LastSeen: lastSeen}
		v, found := db.Expired[k]
		if !found {
			db.Expired[k] = job
			return true
		}
		v.Aux = aux
	} else {
		k := JobKey{Id: id, Host: host}
		v, found := db.Active[k]
		if !found {
			db.Active[k] = job
			return true
		}
		v.LastSeen = lastSeen
		v.Aux = aux
	}
	return false
}

// Purge already-reported jobs from the state if they haven't been seen since before the given
// date, this is to reduce the risk of being confused by jobs whose IDs are reused.

func PurgeJobsBefore(db *JobDatabase, purgeDate time.Time) int {
	active_dead := make([]JobKey, 0)
	expired_dead := make([]ExpiredJobKey, 0)
	deleted := 0
	for k, jobState := range db.Active {
		if jobState.LastSeen.Before(purgeDate) && jobState.IsReported {
			active_dead = append(active_dead, k)
		}
	}
	for _, k := range active_dead {
		delete(db.Active, k)
		deleted++
	}
	for k, jobState := range db.Expired {
		if jobState.LastSeen.Before(purgeDate) && jobState.IsReported {
			expired_dead = append(expired_dead, k)
		}
	}
	for _, k := range expired_dead {
		delete(db.Expired, k)
		deleted++
	}
	return deleted
}

// TODO: It's possible this should sort the output by increasing ID (host then job ID).  This
// basically amounts to creating an array of job IDs, sorting that, and then walking it and looking
// up data by ID when writing.  This is nice because it means that files can be diffed.
//
// TODO: It's possible this should rename the existing state file as a .bak file.

func WriteJobDatabase(stateFilename string, db *JobDatabase) error {
	output_records := make([]map[string]string, 0)
	for _, r := range db.Active {
		output_records = append(output_records, makeMap(r, false))
	}
	for _, r := range db.Expired {
		output_records = append(output_records, makeMap(r, true))
	}
	fields := []string{"id", "host", "startedOnOrBefore", "firstViolation", "lastSeen", "isReported", "isExpired"}
	err := freecsv.WriteFreeCSV(stateFilename, fields, output_records)
	if err != nil {
		return err
	}
	return nil
}

func makeMap(r *JobState, expired bool) map[string]string {
	m := make(map[string]string)
	m["id"] = strconv.FormatUint(uint64(r.Id), 10)
	m["host"] = r.Host
	m["startedOnOrBefore"] = r.StartedOnOrBefore.Format(time.RFC3339)
	m["firstViolation"] = r.FirstViolation.Format(time.RFC3339)
	m["lastSeen"] = r.LastSeen.Format(time.RFC3339)
	m["isReported"] = strconv.FormatBool(r.IsReported)
	m["isExpired"] = strconv.FormatBool(expired)
	return m
}
