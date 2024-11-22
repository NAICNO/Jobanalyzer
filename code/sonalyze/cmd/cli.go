package cmd

import (
	"bufio"
	"cmp"
	"flag"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"

	umaps "go-utils/maps"
)

type CLI struct {
	*flag.FlagSet
	currentGroup   string
	groupForOption map[string]string // option -> group name
}

var (
	// All known options *must* be here but there can be repeated sort values
	priority = map[string]int{
		"application-control": 1,
		"daemon-configuration": 1,
		"data-target": 1,
		"operation-selection": 1,
		"aggregation": 2,
		"job-filter": 3,
		"printing": 4,
		"record-filter": 5,
		"remote-data-source": 6,
		"local-data-source": 7,
		"development": 8,
	}
)

func CLIOutput() io.Writer {
	return flag.CommandLine.Output()
}

func NewCLI(verb string, command Command, name string, exitOnError bool) *CLI {
	fsFlag := flag.ContinueOnError
	if exitOnError {
		fsFlag = flag.ExitOnError
	}
	cli := &CLI{
		FlagSet: flag.NewFlagSet(name, fsFlag),
		groupForOption: make(map[string]string),
	}
	out := CLIOutput()
	cli.FlagSet.Usage = func() {
		restargs := ""
		if _, ok := command.(SetRestArgumentsAPI); ok {
			restargs = " [-- logfile ...]"
		}
		fmt.Fprintf(
			out,
			"Usage: %s %s [options]%s\n\n",
			name,
			verb,
			restargs,
		)
		for _, s := range command.Summary() {
			fmt.Fprintln(out, "  ", s)
		}
		defaults := cli.getSortedDefaults(restargs != "")
		for _, g := range defaults {
			fmt.Fprintf(out, "\n%s options:\n\n", g.group)
			for _, l := range g.text {
				fmt.Fprintln(out, l)
			}
		}
	}
	return cli
}

// Call Group to tag subsequent options with the logical group they belong to, so that when help is
// printed, the options in the same group are presented together.

func (cli *CLI) Group(name string) {
	if _, found := priority[name]; !found {
		panic(fmt.Sprintf("Unknown group %s", name))
	}
	cli.currentGroup = name
}

func (cli *CLI) BoolVar(v *bool, name string, def bool, usage string) {
	cli.tag(name)
	cli.FlagSet.BoolVar(v, name, def, usage)
}

func (cli *CLI) UintVar(v *uint, name string, def uint, usage string) {
	cli.tag(name)
	cli.FlagSet.UintVar(v, name, def, usage)
}

func (cli *CLI) Float64Var(v *float64, name string, def float64, usage string) {
	cli.tag(name)
	cli.FlagSet.Float64Var(v, name, def, usage)
}

func (cli *CLI) StringVar(v *string, name string, def string, usage string) {
	cli.tag(name)
	cli.FlagSet.StringVar(v, name, def, usage)
}

func (cli *CLI) Var(value flag.Value, name string, usage string) {
	cli.tag(name)
	cli.FlagSet.Var(value, name, usage)
}

func (cli *CLI) tag(option string) {
	if cli.currentGroup == "" {
		panic(fmt.Sprintf("No option group set when registering option %s", option))
	}
	if cli.groupForOption[option] != "" {
		panic(fmt.Sprintf("Multiple groups for option %s: %s and %s",
			option, cli.groupForOption[option], cli.currentGroup))
	}
	cli.groupForOption[option] = cli.currentGroup
}

type defaultGroup struct {
	group string
	text []string
}

func (cli *CLI) getSortedDefaults(restArgs bool) []defaultGroup {
	defaultsMap := cli.parseDefaults()
	if restArgs {
		ds, found := defaultsMap["local-data-source"]
		if !found {
			ds.group = "local-data-source"
		}
		ds.text = append(ds.text, "  logfile ...", "\tInput data files")
		defaultsMap["local-data-source"] = ds
	}
	defaults := umaps.Values(defaultsMap)
	slices.SortFunc(defaults, func(a, b defaultGroup) int {
		aPri := priority[a.group]
		bPri := priority[b.group]
		if aPri == bPri {
			return cmp.Compare(a.group, b.group)
		}
		return aPri - bPri
	})
	return defaults
}

// Run PrintDefaults, parse the output, and group the options.  There are other ways of doing this -
// we could just collect the usage strings and format them ourselves - but this ensures consistent
// formatting with minimum fuss.

func (cli *CLI) parseDefaults() map[string]defaultGroup {
	// Collect output from PrintDefaults
	defer cli.FlagSet.SetOutput(flag.CommandLine.Output())
	var tmp strings.Builder
	cli.FlagSet.SetOutput(&tmp)
	cli.FlagSet.PrintDefaults()
	text := tmp.String()

	// Parse the output, grouping together lines that belong to the same option group.
	scanner := bufio.NewScanner(strings.NewReader(text))
	defaults := make(map[string]defaultGroup, 0)
	currentOption := ""
	var optionText []string
	for scanner.Scan() {
		s := scanner.Text()
		if m := optRe.FindStringSubmatch(s); m != nil {
			cli.extendGroup(defaults, currentOption, optionText)
			currentOption = m[1]
			optionText = nil
		}
		optionText = append(optionText, s)
	}
	cli.extendGroup(defaults, currentOption, optionText)
	return defaults
}

func (cli *CLI) extendGroup(
	defaults map[string]defaultGroup,
	currentOption string,
	optionText []string,
) {
	if currentOption != "" {
		group := cli.groupForOption[currentOption]
		if group == "" {
			panic(fmt.Sprintf("No group for option %s", currentOption))
		}
		d, found := defaults[group]
		if !found {
			d.group = group
		}
		d.text = append(d.text, optionText...)
		defaults[group] = d
	}
}

// Brittle!  Wants a test case!
var optRe = regexp.MustCompile(`^  -(\S+)`)

