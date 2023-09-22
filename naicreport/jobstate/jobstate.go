// Data persistence for various subsystems that track job state.
//
// The job information is represented on disk in free CSV form.  This means there's some annoying
// serialization and deserialization work, but the data are textual and structured at the same time,
// and this is better for debugging, resilience, and growth, at least for now.  In the future, maybe
// we'll use a gob instead, or a proper database.

package jobstate

import (
	"os"
	"path"
	"strconv"
	"time"

	"naicreport/storage"
)

// Information about CPU hogs stored in the persistent state.  Other data that are needed for
// generating the report can be picked up from the log data for the job ID.

type JobState struct {
	Id                uint32
	Host              string
	StartedOnOrBefore time.Time
	FirstViolation    time.Time
	LastSeen          time.Time
	IsReported        bool
}

// On the ML nodes, (job#, host) identifies a job uniquely because job#s are not coordinated across
// hosts and no job is cross-host.

type JobKey struct {
	Id   uint32
	Host string
}

// Read the job state from disk and return a parsed and error-checked data structure.  Bogus records
// are silently dropped.
//
// If this returns an error, it is the error returned from storage.ReadFreeCSV, see that for more
// information.  No new errors are generated here.

func ReadJobState(dataPath, filename string) (map[JobKey]*JobState, error) {
	stateFilename := path.Join(dataPath, filename)
	stateCsv, err := storage.ReadFreeCSV(stateFilename)
	if err != nil {
		return nil, err
	}
	state := make(map[JobKey]*JobState)
	for _, repr := range stateCsv {
		success := true
		id := storage.GetUint32(repr, "id", &success)
		host := storage.GetString(repr, "host", &success)
		startedOnOrBefore := storage.GetRFC3339(repr, "startedOnOrBefore", &success)
		firstViolation := storage.GetRFC3339(repr, "firstViolation", &success)
		lastSeen := storage.GetRFC3339(repr, "lastSeen", &success)
		isReported := storage.GetBool(repr, "isReported", &success)
		if !success {
			// Bogus record
			continue
		}
		key := JobKey{id, host}
		state[key] = &JobState{
			Id: id,
			Host: host,
			StartedOnOrBefore: startedOnOrBefore,
			FirstViolation: firstViolation,
			LastSeen: lastSeen,
			IsReported: isReported,
		}
	}
	return state, nil
}

func ReadJobStateOrEmpty(dataPath, filename string) (map[JobKey]*JobState, error) {
	state, err := ReadJobState(dataPath, filename)
	if err == nil {
		return state, nil
	}
	_, isPathErr := err.(*os.PathError)
	if isPathErr {
		return make(map[JobKey]*JobState), nil
	}
	return nil, err
}

// If state does not have the job then add it.  In either case set its LastSeen field to lastSeen.
// Return true if added, false if not.

func EnsureJob(state map[JobKey]*JobState, id uint32, host string,
	started, firstViolation, lastSeen time.Time) bool {
	k := JobKey{Id: id, Host: host}
	v, found := state[k]
	if !found {
		state[k] = &JobState {
			Id: id,
				Host: host,
				StartedOnOrBefore: started,
				FirstViolation: firstViolation,
				LastSeen: lastSeen,
				IsReported: false,
			};
		return true
	}
	v.LastSeen = lastSeen
	return false
}

// Purge already-reported jobs from the state if they haven't been seen since before the given
// date, this is to reduce the risk of being confused by jobs whose IDs are reused.

func PurgeJobsBefore(state map[JobKey]*JobState, purgeDate time.Time) int {
	dead := make([]JobKey, 0)
	for k, jobState := range state {
		if jobState.LastSeen.Before(purgeDate) && jobState.IsReported {
			dead = append(dead, k)
		}
	}
	deleted := 0
	for _, k := range dead {
		delete(state, k)
		deleted++
	}
	return deleted
}

// TODO: It's possible this should sort the output by increasing ID (host then job ID).  This
// basically amounts to creating an array of job IDs, sorting that, and then walking it and looking
// up data by ID when writing.  This is nice because it means that files can be diffed.
//
// TODO: It's possible this should rename the existing state file as a .bak file.

func WriteJobState(dataPath, filename string, data map[JobKey]*JobState) error {
	output_records := make([]map[string]string, 0)
	for _, r := range data {
		m := make(map[string]string)
		m["id"] = strconv.FormatUint(uint64(r.Id), 10)
		m["host"] = r.Host
		m["startedOnOrBefore"] = r.StartedOnOrBefore.Format(time.RFC3339)
		m["firstViolation"] = r.FirstViolation.Format(time.RFC3339)
		m["lastSeen"] = r.LastSeen.Format(time.RFC3339)
		m["isReported"] = strconv.FormatBool(r.IsReported)
		output_records = append(output_records, m)
	}
	fields := []string{"id", "host", "startedOnOrBefore", "firstViolation", "lastSeen", "isReported"}
	stateFilename := path.Join(dataPath, filename)
	err := storage.WriteFreeCSV(stateFilename, fields, output_records)
	if err != nil {
		return err
	}
	return nil
}
