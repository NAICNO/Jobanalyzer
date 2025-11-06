package special

// Since the database can be opened on various storage media (a proper database, a directory tree, a
// file list) and not all media can supply all data types (for example a file list supplies only a
// subset of the types, by design), the various consumers of data must be careful to request the
// data type they need, and the database must be careful to record the type of the data it can
// provide.
//
// Since most file types in a directory tree representation can supply multiple types of data,
// however, the type bits can be OR'ed together.  A sysinfo file can supply both node and card data;
// a sample file can supply process, cpu, and gpu sample; a cluzter file can supply slurm cluster,
// node, and partition data.
//
// The data type is independent of the representation of the data (CSV, JSON, whatever), which is
// why the special.DataType bits are not the same as the filedb.FileAttr bits.

type DataType int

const (
	// Primitive data types, usually what consumers request.
	SampleData DataType = 1 << iota
	NodeSampleData
	CpuSampleData
	GpuSampleData
	NodeData
	CardData
	SlurmJobData
	SlurmNodeData
	SlurmPartitionData
	SlurmClusterData

	// Canonical sets of types: these are the data that are represented together in the same file,
	// in a file-based store.
	ProcessSampleData = SampleData | NodeSampleData | CpuSampleData | GpuSampleData
	SysinfoData       = NodeData | CardData
	SlurmSystemData   = SlurmNodeData | SlurmPartitionData | SlurmClusterData
)
