/// The cached representation is a structured file:
///
/// File prefix
/// String pool
/// Field offsets
/// LogEntries
///
/// The file prefix:
///
///   magic:   u32      0x12045078, checks endianness too
///   version: u32      version number
///   size:    u16      size of LogEntry in bytes
///   strings: u32      number of strings in string pool
///   stringbytes: u32  number of bytes in string pool
///   fields:  u16      number of fields in LogEntry
///   fieldbytes: u32   number of bytes in field offset pool
///   logentries: u32   number of logentries in logentry pool
///   logbytes:         number of bytes in the logentry pool
///
/// The string pool: each entry looks like this
///
///   length: u32       number of bytes
///   bytes: [u8; length] the bytes
///
/// The field offset pool: each entry looks like this
///
///   name: u32         index in string pool
///   type: u32         index in string pool, the name of the type (eg "u32")
///   offset: u32       offset within logentry
///
/// The LogEntries: each entry is one of the CachedLogEntry structures below.
///
/// To use a cached file it must be validated:
///
/// - the corresponding csv file must exist
/// - the cached modification time must be strictly after the modification time of the
///   corresponding csv
/// - the cached file must have the right magic number (endianness) and compatible version
/// - the cached file must have a size that corresponds to the sum of the header, the
///   sum of the size fields for the different sections
/// - when read, each item must be sensible and not exceed the pool
/// - the struct layout in the file must match the struct layout in the program: field names,
///   order, offsets.

/// TODO: Ideally we have a format that allows streaming reading and streaming writing; reading is
/// probably OK but writing is trickier since the header needs a ton of information - though
/// fortunately not file offsets.

/// This has been packed manually.  If there's a directive to say "don't fuck with this" we should use it.
/// TODO: Assert sensible size
pub struct CachedLogEntry {
    pub timestamp: i64,
    pub mem_gb: f64,
    pub gpumem_gb: f64,
    pub cputime_sec: f64,
    pub gpus: u64,              // this is ~0 for "unknown" otherwise a bitvector
    pub hostname: u32,          // string pool index
    pub memtotal_gb: f32,
    pub user: u32,              // string pool index
    pub pid: u32,
    pub job_id: u32,
    pub command: u32,           // string pool index
    pub cpu_pct: f32,
    pub rssanon_gb: f32,
    pub gpu_pct: f32,
    pub gpumem_pct: f32,
    pub rolledup: u32,
    pub cpu_util_pct: f32,
    pub major: u16,
    pub minor: u16,
    pub bugfix: u16,
    pub num_cores: u16,
    pub gpu_status: u8,         // Some translation
    pub padding: [u8; 7],
}

/// Following https://stackoverflow.com/questions/28127165/how-to-convert-struct-to-u8, we
/// can copy values into this struct and then treat it as a memory block and write it into
/// some stream.
///
/// TODO: Padding
