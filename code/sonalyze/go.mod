module sonalyze

go 1.23.8

toolchain go1.24.2

require go-utils v0.0.0-00010101000000-000000000000

require github.com/lars-t-hansen/ini v0.3.0

require github.com/NordicHPC/sonar/util/formats v0.0.0-00010101000000-000000000000

replace go-utils => ../go-utils

replace github.com/NordicHPC/sonar/util/formats => ../../../sonar-fmt/util/formats
