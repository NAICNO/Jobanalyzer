module sonalyze

go 1.24.2

require (
	github.com/NordicHPC/sonar/util/formats v0.18.3
	github.com/jackc/pgx/v5 v5.7.6
	github.com/lars-t-hansen/ini v0.3.0
	github.com/twmb/franz-go v1.19.1
	go-utils v0.0.0-00010101000000-000000000000
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	github.com/twmb/franz-go/pkg/kmsg v1.11.2 // indirect
	golang.org/x/crypto v0.38.0 // indirect
	golang.org/x/text v0.25.0 // indirect
)

replace go-utils => ../go-utils
