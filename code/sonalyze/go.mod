module sonalyze

go 1.23.8

toolchain go1.24.2

require go-utils v0.0.0-00010101000000-000000000000

require github.com/lars-t-hansen/ini v0.3.0

require github.com/twmb/franz-go v1.19.1

// This is supposed to be v0.13.0 but I'm unable to coerce the Go
// module system into honoring that.
require github.com/NordicHPC/sonar/util/formats v0.0.0-20250516095736-4f3fa9614148

require (
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/twmb/franz-go/pkg/kmsg v1.11.2 // indirect
)

replace go-utils => ../go-utils
