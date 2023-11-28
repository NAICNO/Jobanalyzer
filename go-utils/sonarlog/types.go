package sonarlog

type SonarReading struct {
	Version    string  `json:"version"`
	Timestamp  string  `json:"time"`
	Host       string  `json:"host"`
	Cores      uint64  `json:"cores"`
	User       string  `json:"user"`
	Job        uint64  `json:"job,omitempty"`
	Pid        uint64  `json:"pid,omitempty"`
	Cmd        string  `json:"cmd"`
	CpuPct     float64 `json:"cpu%,omitempty"`
	CpuKib     uint64  `json:"cpukib,omitempty"`
	Gpus       string  `json:"gpus,omitempty"`
	GpuPct     float64 `json:"gpu%,omitempty"`
	GpuMemPct  float64 `json:"gpumem%,omitempty"`
	GpuKib     uint64  `json:"gpukib,omitempty"`
	CpuTimeSec uint64  `json:"cputime_sec,omitempty"`
}

type SonarHeartbeat struct {
	Version   string `json:"version"`
	Timestamp string `json:"time"`
	Host      string `json:"host"`
}

type ExfilEnvelope struct {
	Version string `json:"version"`
	Type    string `json:"type"`
	Cluster string `json:"cluster"`
	Host    string `json:"host"`
	Value   []any  `json:"value"`
}
