// Data representation of Slurm job data.
//
// The data representation produced here is a 2-dimensional volume:
//
//   job-attr ...
//   ...
//
// where most job-attrs are primitive.  However, NodeList and ReqGPUS have additional structure and
// represent a hidden third dimension and/or variable-length field and/or variable number of
// columns, depending on your point of view.

package repr

import (
	"unsafe"

	. "sonalyze/common"
)

// Representation of Slurm data (mostly, and historically, from sacct polling).
//
//   - See the sacct documentation for interpretation, we follow that except as noted here, eg timeout
//     has been translated from minutes to seconds.
//
//   - the JobIDRaw field is split here into JobID (integer) and JobStep (string), with the latter
//     being empty for the "main" record for the job.  For array jobs, JobID is parsed into
//     ArrayJobID, ArrayIndex, and ArrayStep.  For het jobs, JobID is parsed into HetJobID,
//     HetOffset, and HetStep.  For normal jobs, the array and het fields are zero/blank.
//
//   - 2^32-1 seconds is about 136 years; it seems like a long time and is fine for elapsed/real time.
//     But 170K cores (Betzy) running flat out for a week comes to about 24 times that.  So fields for
//     total consumed CPU time must be 64 bits.
//
//   - No doubt something can be made of sub-gigabyte memory sizes, but everything here is rounded up
//     to the nearest GB.
//
// - I/O is also presented in GB (anything less isn't meaningful), rounded up to nearest GB.
//
//   - The state field has been stripped of extraneous information, eg, "CANCELLED by ..." is just
//     CANCELLED.
//
//   - For jobs that were cancelled before they got to be scheduled, Start can be 0 and NodeList can
//     be the empty string, and probably a number of other fields are off too in this case.
//
// This structure is unreasonably large, but in practice there are many fewer of these (several
// orders of magnitude fewer) than the Sonar sample records.
//
// This structure is essentially a flattened view onto newfmt.SlurmJob where newfmt is
// newfmt=github.com/NordicHPC/sonar/util/formats/newfmt.  The reason it is a separate view is
// partly the flattening and partly the desire to convert string to Ustr and uint64 to uint32, so as
// to make this a pointer-free data structure - since there can be many of them and it is large.
type SacctInfo struct {
	Time         int64  // Unix time - record timestamp, not in older data
	Start        int64  // Unix time
	End          int64  // Unix time
	Submit       int64  // Unix time
	SystemCPU    uint64 // seconds of cpu time
	UserCPU      uint64 // seconds of cpu time
	AveCPU       uint64 // seconds of cpu time
	MinCPU       uint64 // seconds of cpu time
	Version      Ustr
	User         Ustr // only for the "main" record for the job
	JobName      Ustr
	State        Ustr // uppercase string, eg COMPLETED, TIMEOUT
	Account      Ustr
	Layout       Ustr
	Reservation  Ustr
	JobStep      Ustr // name of step if any, eg "extern" or "1"
	ArrayStep    Ustr
	HetStep      Ustr
	NodeList     Ustr // compressed nodelist, for now, though this could be problematic
	Partition    Ustr
	ReqGPUS      Ustr // comma-separated list of model=n and/or *=n from AllocTRES field
	JobID        uint32
	ArrayJobID   uint32
	ArrayIndex   uint32
	HetJobID     uint32
	HetOffset    uint32
	AveDiskRead  uint32 // GB
	AveDiskWrite uint32 // GB
	AveRSS       uint32 // GB
	AveVMSize    uint32 // GB
	ElapsedRaw   uint32 // seconds of real time
	MaxRSS       uint32 // GB
	MaxVMSize    uint32 // GB
	ReqCPUS      uint32
	ReqMem       uint32 // GB
	ReqNodes     uint32
	Suspended    uint32 // seconds of real time
	TimelimitRaw uint32 // *seconds* of real time (input data has minutes)
	ExitCode     uint8  // the code part of code:signal
	ExitSignal   uint8  // the signal part of code:signal - only in older data
}

var (
	// MT: Constant after initialization; immutable
	SizeofSacctInfo uintptr
)

func init() {
	var x SacctInfo
	SizeofSacctInfo = unsafe.Sizeof(x)
}
