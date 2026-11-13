module sonalyze

go 1.26.0

require (
	github.com/NordicHPC/sonar/util/formats v0.18.4
	github.com/danielgtaylor/huma/v2 v2.40.0
	github.com/jackc/pgx/v5 v5.13.0
	github.com/klauspost/compress v1.20.4
	github.com/lars-t-hansen/ini v0.3.0
	github.com/twmb/franz-go v1.22.3
	go-utils v0.0.0-00010101000000-000000000000
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/pierrec/lz4/v4 v4.1.33 // indirect
	github.com/twmb/franz-go/pkg/kmsg v1.15.1 // indirect
	golang.org/x/text v0.44.0 // indirect
)

replace go-utils => ../go-utils
