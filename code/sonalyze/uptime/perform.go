// Given a stretch of time - a set of Samples - when a host was up, the status of its GPUs can be
// determined by looking at the records' gpu_status fields.
//
// In addition to the Samples, we take as inputs the `from` and `to` timestamps defining the time
// window of interest.  A host is up at the start if its first Sample is within the gap-threshold
// of the `from` time, and ditto it is up at the end for its last Sample close to the `to` time.
// The gap-threshold is computed from the sampling interval provided as an argument to the program.
//
// The output has five fields: device, host, state, start, end where
//
//  - device is `host` or `gpu`
//  - host is the name of the host (FQDN probably)
//  - state is `up` or `down`
//  - start is the inclusive start of the window when the device was in the given state, on the form
//    YYYY-MM-DD HH:MM (the same form used elsewhere)
//  - end is the exclusive end of the window, ditto
//
// `start` and `end` of hosts are computed so that windows overlap: the `end` of one record will
// equal the `start` of the next.  This is fine, and helps clients display the data.  `start` and
// `end` of gpus similarly form a complete timeline within the time that its host is up.
//
// For csvnamed the field names are as given above, and all the values are strings.
//
// Outputs are sorted by host name and then increasing time of the `start` field.  This means the
// report can be read top-to-bottom to get a chronological sense for the history of each host.
//
// FUTURE:
//
// - For nodes/hosts that don't have GPUs it would be nice not to print any GPU information.
//   We should be able to use the config data to drive that.
//
// - (Speculative) At the moment, the gpu_status is per-host, not per-card, because that's all sonar
//   is able to discern.  When that changes, the device field will be generalized so that its value
//   may be `gpu0`, `gpu1`, etc.  Most likely records for these will be in addition to the records
//   for plain `gpu`, which will plausibly retain its existing semantics.
//
// - (Speculative) As gpu_status is an enum it can take on other values than up or down; thus when
//   we improve state detection, the representation here of that value may change, or there may be
//   additional fields.

package uptime

import (
	"io"
	"sort"

	"go-utils/config"
	"go-utils/hostglob"
	"go-utils/maps"
	"go-utils/slices"
	. "sonalyze/common"
	"sonalyze/db"
	"sonalyze/sonarlog"
)

// The time ranges are not clearly inclusive or exclusive, it depends on the data available.  We use
// the inclusive end points of the total date range when we're at the ends of that range, and the
// dates from a boundary Sample otherwise.  Adjacent reports will tend to overlap, by design, as the
// date used for the end of one report is the same as the one for the beginning of the next.
//
// The start and end times are represented as strings to be compatible with the Rust code, which
// rounds down the time to minute precision.  This matters only when there are some closely related
// records in the data stream, which normally does not happen but can happen, and since the printing
// code only prints times to minute precision the data would appear unsorted on output if we were to
// use second precision internally.
//
// TODO: IMPROVEME: It's a bug in both the Rust code and this code that the folding of the timestamp
// to string (and the rounding off to minute precision) happens *after* we determine the timeline.
// In some cases, this will result in duplicate records in the output, where the data stream we
// operated on had sub-minute precision and we constructed records according to that.  Mostly this
// will happen when the data are somewhat wonky, but it's still wrong.
type report struct {
	device     string
	host       string
	state      string
	start, end string
}

// TODO: CLEANUP: The window is just a []sonarlog.Sample, the indices are a Rust-ism.
type window struct {
	start, end int // inclusive indices in `samples`
}

func (uc *UptimeCommand) Perform(
	out io.Writer,
	cfg *config.ClusterConfig,
	_ db.SampleCluster,
	streams sonarlog.InputStreamSet,
	bounds sonarlog.Timebounds,
	hostGlobber *hostglob.HostGlobber,
	_ func(*sonarlog.Sample) bool,
) error {
	samples := slices.CatenateP(maps.Values(streams))
	if uc.Verbose {
		Log.Infof("%d streams", len(streams))
		Log.Infof("%d records after hack", len(samples))
	}
	uc.printReports(out, uc.computeReports(samples, bounds, cfg, hostGlobber))
	return nil
}

// Compute up/down reports for all selected hosts within the time window.  The result will not be
// sorted by anything.

func (uc *UptimeCommand) computeReports(
	samples sonarlog.SampleStream,
	bounds sonarlog.Timebounds,
	cfg *config.ClusterConfig,
	hostGlobber *hostglob.HostGlobber,
) []report {
	reports := make([]report, 0)
	fromIncl, toIncl := uc.InterpretFromToWithBounds(bounds)

	sort.Stable(sonarlog.HostTimeSortableSampleStream(samples))

	uc.computeAlwaysDown(&reports, samples, cfg, hostGlobber, fromIncl, toIncl)

	hostUpWindows := make([]window, 0)
	cutoff := int64(uc.Interval) * 60 * 2
	for _, w := range uc.computeHostWindows(samples, hostGlobber, fromIncl, toIncl) {
		hostFirst := samples[w.start]
		hostLast := samples[w.end]

		// If the host is down at the start, push out a record saying so.  Then we start in the "up"
		// state always.
		if !(hostFirst.S.Timestamp-fromIncl <= cutoff) {
			if uc.Verbose {
				Log.Infof("  Down at start")
			}
			if !uc.OnlyUp {
				reports = append(reports, report{
					device: "host",
					host:   hostFirst.S.Host.String(),
					state:  "down",
					start:  formatTime(fromIncl),
					end:    formatTime(hostFirst.S.Timestamp),
				})
			}
		}

		// If the host is down at the end, push out a record saying so.
		if !(toIncl-hostLast.S.Timestamp <= cutoff) {
			if uc.Verbose {
				Log.Infof("  Down at end")
			}
			if !uc.OnlyUp {
				reports = append(reports, report{
					device: "host",
					host:   hostFirst.S.Host.String(),
					state:  "down",
					start:  formatTime(hostLast.S.Timestamp),
					end:    formatTime(toIncl),
				})
			}
		}

		// Within the relevant window of the host's entries, we need to figure out when it might
		// have been down and push out down/up records.  It will be up at the beginning and end,
		// we've ensured that above.
		windowStart := w.start
		for {
			prevTimestamp := samples[windowStart].S.Timestamp

			// We're in an "up" window, scan to its end.
			j := windowStart + 1
			for j <= w.end && samples[j].S.Timestamp-prevTimestamp <= cutoff {
				prevTimestamp = samples[j].S.Timestamp
				j++
			}

			// Now j points past the last record in the up window.  There's a chance here that start
			// and end are the same value (only one sample between two "down" windows); nothing to
			// be done about that.
			if uc.Verbose {
				Log.Infof("  Up window %d..%d inclusive", windowStart, j-1)
			}
			if !uc.OnlyDown {
				reports = append(reports, report{
					device: "host",
					host:   hostFirst.S.Host.String(),
					state:  "up",
					start:  formatTime(samples[windowStart].S.Timestamp),
					end:    formatTime(samples[j-1].S.Timestamp),
				})
			}

			// Record this window, we'll need it for the GPU scans later.  (The scans could happen
			// here, but it just makes the code unreadable.)
			hostUpWindows = append(hostUpWindows, window{windowStart, j - 1})

			if j > w.end {
				break
			}

			// System went down in the window.  The window in which it is down is entirely between
			// these two records.  The fact that there is a following record means it came up again.
			if uc.Verbose {
				Log.Infof("  Down window %d..%d inclusive\n", j-1, j)
			}
			if !uc.OnlyUp {
				reports = append(reports, report{
					device: "host",
					host:   hostFirst.S.Host.String(),
					state:  "down",
					start:  formatTime(prevTimestamp),
					end:    formatTime(samples[j].S.Timestamp),
				})
			}

			windowStart = j
		}
	}

	// Now for each host "up" window, figure out the GPU status within that window.
	for _, w := range hostUpWindows {
		i := w.start
		for i <= w.end {
			gpuIsUp := samples[i].S.GpuFail == 0
			start := i
			for i <= w.end && (samples[i].S.GpuFail == 0) == gpuIsUp {
				i++
			}
			updown := "down"
			if gpuIsUp {
				updown = "up"
			}
			if !(updown == "up" && uc.OnlyDown) && !(updown == "down" && uc.OnlyUp) {
				reports = append(reports, report{
					device: "gpu",
					host:   samples[w.start].S.Host.String(),
					state:  updown,
					start:  formatTime(samples[start].S.Timestamp),
					end:    formatTime(samples[min(w.end, i)].S.Timestamp),
				})
			}
		}
	}

	return reports
}

// Return a sequence of windows: each window pertains to a stretch of records for a single host
// starting no earlier than fromIncl and ending no later than toIncl.
//
// Samples are sorted by host and then timestamp.  Therefore there is at most one window per host of
// interest.  The host name is given by the Host field of the first record in the window.

func (uc *UptimeCommand) computeHostWindows(
	samples sonarlog.SampleStream,
	hostGlobber *hostglob.HostGlobber,
	fromIncl, toIncl int64,
) []window {
	windows := make([]window, 0)
	i := 0
	lim := len(samples)
	for i < lim {
		// Skip anything for before the window we're interested in.
		for i < lim && samples[i].S.Timestamp < fromIncl {
			i++
		}
		if i == lim {
			break
		}
		// Collect the window
		hostStart := i
		hostEnd := i
		host := samples[hostStart].S.Host
		hostStr := host.String()
		i++
		for i < lim && samples[i].S.Host == host {
			if samples[i].S.Timestamp <= toIncl {
				hostEnd = i
			}
			i++
		}
		// If the host is excluded, we'll skip it
		if !hostGlobber.IsEmpty() && !hostGlobber.Match(hostStr) {
			continue
		}

		// We have an included host and a window.
		if uc.Verbose {
			Log.Infof("%s: %d..%d inclusive, i=%d", host, hostStart, hostEnd, i)
		}

		windows = append(windows, window{hostStart, hostEnd})
	}
	return windows
}

// If there is a cluster config then this induces a set of host names.  If there is a host in that
// set that is not in the list of samples *and* which passes the host filter, then that host is down
// for the entire window.
//
// The samples are sorted by host and then by ascending timestamp.

func (uc *UptimeCommand) computeAlwaysDown(
	reports *[]report,
	samples sonarlog.SampleStream,
	cfg *config.ClusterConfig,
	hostGlobber *hostglob.HostGlobber,
	fromIncl, toIncl int64,
) {
	if !uc.OnlyUp && cfg != nil {
		hs := make(map[Ustr]bool)
		for _, h := range cfg.HostsDefinedInTimeWindow(fromIncl, toIncl) {
			hs[StringToUstr(h)] = true
		}
		for _, sample := range samples {
			delete(hs, sample.S.Host)
		}
		for h := range hs {
			if !hostGlobber.IsEmpty() && !hostGlobber.Match(h.String()) {
				continue
			}
			// `h` is down in the entire window
			*reports = append(*reports, report{
				device: "host",
				host:   h.String(),
				state:  "down",
				start:  formatTime(fromIncl),
				end:    formatTime(toIncl),
			})
		}
	}
}

func formatTime(t int64) string {
	return FormatYyyyMmDdHhMmUtc(t)
}
