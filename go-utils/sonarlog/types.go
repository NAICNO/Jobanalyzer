package sonarlog

// Version 0.5.0 had fields through CpuKib
//
// Version 0.6.0 added GPU fields
//
// Version 0.7.0 added CpuTimeSec and Rolledup, and mostly deprecated CpuPct b/c it is
// percent-since-start, not since last measurement
//
// Cluster is implied in sonar data, it is added during json encoding

type SonarReading struct {
	Version    string  `json:"version"`  // semver
	Timestamp  string  `json:"time"`     // Full ISO timestamp: yyyy-mm-ddThh:mm:ss+hh:mm
	Cluster    string  `json:"cluster"`
	Host       string  `json:"host"`
	Cores      uint64  `json:"cores"`
	User       string  `json:"user"`
	Job        uint64  `json:"job,omitempty"`
	Pid        uint64  `json:"pid,omitempty"`
	Cmd        string  `json:"cmd"`
	CpuPct     float64 `json:"cpu%,omitempty"`
	CpuKib     uint64  `json:"cpukib,omitempty"`
	Gpus       string  `json:"gpus,omitempty"` // unknown, none, or k,n,m,...
	GpuPct     float64 `json:"gpu%,omitempty"`
	GpuMemPct  float64 `json:"gpumem%,omitempty"`
	GpuKib     uint64  `json:"gpukib,omitempty"`
	CpuTimeSec uint64  `json:"cputime_sec,omitempty"`
	Rolledup   uint64  `json:"rolledup,omitempty"`
}

type SonarHeartbeat struct {
	Version   string `json:"version"`
	Timestamp string `json:"time"`
	Cluster    string  `json:"cluster"`
	Host      string `json:"host"`
}
