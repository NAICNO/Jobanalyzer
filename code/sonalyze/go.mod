module sonalyze

go 1.25.13

require (
	github.com/NordicHPC/sonar/util/formats v0.18.4
	github.com/danielgtaylor/huma/v2 v2.39.1
	github.com/jackc/pgx/v5 v5.9.2
	github.com/klauspost/compress v1.19.2
	github.com/lars-t-hansen/ini v0.3.0
	github.com/twmb/franz-go v1.21.7
	go-utils v0.0.0-00010101000000-000000000000
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/pierrec/lz4/v4 v4.1.26 // indirect
	github.com/twmb/franz-go/pkg/kmsg v1.13.1 // indirect
	golang.org/x/text v0.40.0 // indirect
)

replace go-utils => ../go-utils
