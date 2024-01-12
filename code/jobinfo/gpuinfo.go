package main

const (
	noGpuDevice = -1
	invalidUID  = 6666666
)

type gpuInfo struct {
	device     int     // Device ID or noGpuDevice
	pid        uint    // Process ID
	user       string  // User name, _zombie_PID for zombies
	uid        uint    // User ID, invalidUID for zombies
	gpuPct     float64 // Percent of GPU /for this sample/, 0.0 for zombies
	memPct     float64 // Percent of memory /for this sample/, 0.0 for zombies
	memSizeKiB uint64  // Memory use in KiB /for this sample/, _not_ zero for zombies
	command    string  // The command, _unknown_ for zombies, _noinfo_ if not known
}
