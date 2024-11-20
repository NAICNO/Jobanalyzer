package application

import (
	"fmt"
	"io"

	"sonalyze/cmd"
	"sonalyze/cmd/add"
	"sonalyze/cmd/clusters"
	"sonalyze/cmd/configs"
	"sonalyze/cmd/jobs"
	"sonalyze/cmd/load"
	"sonalyze/cmd/metadata"
	"sonalyze/cmd/nodes"
	"sonalyze/cmd/parse"
	"sonalyze/cmd/profile"
	"sonalyze/cmd/sacct"
	"sonalyze/cmd/top"
	"sonalyze/cmd/uptime"
)

func CommandHelp(out io.Writer) {
	fmt.Fprintf(out, "  add      - add data to the database\n")
	fmt.Fprintf(out, "  cluster  - print cluster information\n")
	fmt.Fprintf(out, "  config   - print node information extracted from cluster config\n")
	fmt.Fprintf(out, "  jobs     - summarize and filter jobs\n")
	fmt.Fprintf(out, "  load     - print system load across time\n")
	fmt.Fprintf(out, "  metadata - parse data, print stats and metadata\n")
	fmt.Fprintf(out, "  node     - print node information extracted from sysinfo table\n")
	fmt.Fprintf(out, "  profile  - print the profile of a particular job\n")
	fmt.Fprintf(out, "  sacct    - print information extracted from Slurm sacct data\n")
	fmt.Fprintf(out, "  sample   - print sonar sample information (aka `parse`)\n")
	fmt.Fprintf(out, "  top      - print per-cpu load information across time\n")
	fmt.Fprintf(out, "  uptime   - print aggregated information about system uptime\n")
	fmt.Fprintf(out, "  version  - print information about the program\n")
	fmt.Fprintf(out, "  help     - print this message\n")
}

func ConstructCommand(verb string) (command cmd.Command, actualVerb string) {
	switch verb {
	case "add":
		command = new(add.AddCommand)
	case "cluster":
		command = new(clusters.ClusterCommand)
	case "config":
		command = new(configs.ConfigCommand)
	case "node":
		command = new(nodes.NodeCommand)
	case "jobs":
		command = new(jobs.JobsCommand)
	case "load":
		command = new(load.LoadCommand)
	case "meta", "metadata":
		command = new(metadata.MetadataCommand)
		verb = "metadata"
	case "sample", "parse":
		command = new(parse.ParseCommand)
		verb = "sample"
	case "profile":
		command = new(profile.ProfileCommand)
	case "sacct":
		command = new(sacct.SacctCommand)
	case "top":
		command = new(top.TopCommand)
	case "uptime":
		command = new(uptime.UptimeCommand)
	}
	actualVerb = verb
	return
}
