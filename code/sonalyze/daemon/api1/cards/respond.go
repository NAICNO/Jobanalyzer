// Generated from cards.go by generate-response.  DO NOT EDIT.

package cards

import (
	"sonalyze/daemon/apiutil"
	"sonalyze/db/repr"
)

const responseDefaults = "Time,Node,Manufacturer,Model,Memory"

type Cards_Card struct {
	Time           string `json:"Time,omitempty" doc:"Full ISO timestamp of when the reading was taken"`
	Node           string `json:"Node,omitempty" doc:"Card's node at this time"`
	Index          uint64 `json:"Index,omitempty" doc:"Card's index on its node at this time"`
	UUID           string `json:"UUID,omitempty" doc:"Card's unique identifier (but not necessarily its only unique identifier)"`
	Address        string `json:"Address,omitempty" doc:"Card's address on its node at this time"`
	Manufacturer   string `json:"Manufacturer,omitempty" doc:"Card's manufacturer's name"`
	Model          string `json:"Model,omitempty" doc:"Card model"`
	Architecture   string `json:"Architecture,omitempty" doc:"Card's architecture name"`
	Driver         string `json:"Driver,omitempty" doc:"Card driver's version at this time"`
	Firmware       string `json:"Firmware,omitempty" doc:"Card firmware's version at this time"`
	Memory         uint64 `json:"Memory,omitempty" doc:"Card's memory in KB"`
	PowerLimit     uint64 `json:"PowerLimit,omitempty" doc:"Card's power limit at this time"`
	MaxPowerLimit  uint64 `json:"MaxPowerLimit,omitempty" doc:"Card's maximum power limit"`
	MinPowerLimit  uint64 `json:"MinPowerLimit,omitempty" doc:"Card's minimum power limit"`
	MaxCEClock     uint64 `json:"MaxCEClock,omitempty" doc:"Card's maximum compute element clock speed"`
	MaxMemoryClock uint64 `json:"MaxMemoryClock,omitempty" doc:"Card's maximum memory clock speed"`
}

func respond(flds *apiutil.FieldMap, r *repr.SysinfoCardData) Cards_Card {
	var x Cards_Card
	if flds.Has("Time") {
		x.Time = r.Time
	}
	if flds.Has("Node") {
		x.Node = r.Node
	}
	if flds.Has("Index") {
		x.Index = r.Index
	}
	if flds.Has("UUID") {
		x.UUID = r.UUID
	}
	if flds.Has("Address") {
		x.Address = r.Address
	}
	if flds.Has("Manufacturer") {
		x.Manufacturer = r.Manufacturer
	}
	if flds.Has("Model") {
		x.Model = r.Model
	}
	if flds.Has("Architecture") {
		x.Architecture = r.Architecture
	}
	if flds.Has("Driver") {
		x.Driver = r.Driver
	}
	if flds.Has("Firmware") {
		x.Firmware = r.Firmware
	}
	if flds.Has("Memory") {
		x.Memory = r.Memory
	}
	if flds.Has("PowerLimit") {
		x.PowerLimit = r.PowerLimit
	}
	if flds.Has("MaxPowerLimit") {
		x.MaxPowerLimit = r.MaxPowerLimit
	}
	if flds.Has("MinPowerLimit") {
		x.MinPowerLimit = r.MinPowerLimit
	}
	if flds.Has("MaxCEClock") {
		x.MaxCEClock = r.MaxCEClock
	}
	if flds.Has("MaxMemoryClock") {
		x.MaxMemoryClock = r.MaxMemoryClock
	}
	return x
}
