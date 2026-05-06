// Generated from cards.go by generate-response.  DO NOT EDIT.

package cards

import (
	"sonalyze/daemon/apiutil"
	"sonalyze/db/repr"
)

const responseDefaults = "Time,Node,Manufacturer,Model,Memory"

type Card_Card struct {
	Time           string `json:"Time,omitempty"`
	Node           string `json:"Node,omitempty"`
	Index          uint64 `json:"Index,omitempty"`
	UUID           string `json:"UUID,omitempty"`
	Address        string `json:"Address,omitempty"`
	Manufacturer   string `json:"Manufacturer,omitempty"`
	Model          string `json:"Model,omitempty"`
	Architecture   string `json:"Architecture,omitempty"`
	Driver         string `json:"Driver,omitempty"`
	Firmware       string `json:"Firmware,omitempty"`
	Memory         uint64 `json:"Memory,omitempty"`
	PowerLimit     uint64 `json:"PowerLimit,omitempty"`
	MaxPowerLimit  uint64 `json:"MaxPowerLimit,omitempty"`
	MinPowerLimit  uint64 `json:"MinPowerLimit,omitempty"`
	MaxCEClock     uint64 `json:"MaxCEClock,omitempty"`
	MaxMemoryClock uint64 `json:"MaxMemoryClock,omitempty"`
}

func respond(flds *apiutil.FieldMap, r *repr.SysinfoCardData) Card_Card {
	var x Card_Card
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
