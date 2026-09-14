// Generated from cardprof.go by generate-response.  DO NOT EDIT.

package cardprof

import (
	. "sonalyze/cmd/gpus"
	"sonalyze/daemon/apiutil"
	. "sonalyze/table"
)

const responseDefaults = "Timestamp,Hostname,Memory,Power"

type CardProfile_Timestep struct {
	Timestamp   string `json:"Timestamp,omitempty" doc:"Timestamp of when the reading was taken"`
	Hostname    string `json:"Hostname,omitempty" doc:"Name that host is known by on the cluster"`
	Index       uint64 `json:"Index,omitempty" doc:"Card index on the host"`
	Fan         uint64 `json:"Fan,omitempty" doc:"Fan speed in percent of max"`
	Memory      uint64 `json:"Memory,omitempty" doc:"Amount of memory in use"`
	Temperature int64  `json:"Temperature,omitempty" doc:"Card temperature in degrees C"`
	Power       uint64 `json:"Power,omitempty" doc:"Current power draw in Watts"`
	PowerLimit  uint64 `json:"PowerLimit,omitempty" doc:"Current power limit in Watts"`
	CEClock     uint64 `json:"CEClock,omitempty" doc:"Current compute element clock in MHz"`
	MemoryClock uint64 `json:"MemoryClock,omitempty" doc:"Current memory clock in MHz"`
}

func respond(flds *apiutil.FieldMap, r *ReportLine) CardProfile_Timestep {
	var x CardProfile_Timestep
	if flds.Has("Timestamp") {
		x.Timestamp = JSONFromDateTimeValue(r.Timestamp)
	}
	if flds.Has("Hostname") {
		x.Hostname = JSONFromUstr(r.Hostname)
	}
	if flds.Has("Index") {
		x.Index = r.Index
	}
	if flds.Has("Fan") {
		x.Fan = r.Fan
	}
	if flds.Has("Memory") {
		x.Memory = r.Memory
	}
	if flds.Has("Temperature") {
		x.Temperature = r.Temperature
	}
	if flds.Has("Power") {
		x.Power = r.Power
	}
	if flds.Has("PowerLimit") {
		x.PowerLimit = r.PowerLimit
	}
	if flds.Has("CEClock") {
		x.CEClock = r.CEClock
	}
	if flds.Has("MemoryClock") {
		x.MemoryClock = r.MemoryClock
	}
	return x
}
