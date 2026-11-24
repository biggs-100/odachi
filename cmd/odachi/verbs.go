package main

import "strings"

// verb is one command the binary dispatches. Usage holds the exact lines the
// command occupies in the usage block, with their leading two spaces; Example
// is the short form the unknown-command error prints. Both are kept beside the
// name so the help and the error are two renderings of one list (0360).
type verb struct {
	name    string
	usage   []string
	example string
}

// verbCatalog is the one list of commands. Its order is the order of the usage
// block, and both the help and the unknown-command error read from it. The line
// `odachi   the same as odachi start` describes the bare invocation, not a
// command, and stays in usage() outside this table (0360).
var verbCatalog = []verb{
	{
		name: "start",
		usage: []string{
			`  odachi start [flags]`,
			`  odachi start --light "<requirement>"`,
		},
		example: "odachi start",
	},
	{
		name: "ask",
		usage: []string{
			`  odachi ask "<question>"`,
			`  odachi ask --answer "<answer>" [--for P1]`,
			`  odachi ask --craft "<decision>" --because "<reason>"`,
			`  odachi ask --change "<words>" --req R1`,
		},
		example: `odachi ask "<question>"`,
	},
	{
		name: "plan",
		usage: []string{
			`  odachi plan "<text>" [--req R1] [--route inline | delegated --agent <name> --skill <name>]`,
			`  odachi plan --done T2 [--commit abc1234]`,
		},
		example: `odachi plan "<text>" | odachi plan --done T2`,
	},
	{
		name: "delegate",
		usage: []string{
			`  odachi delegate <taskdoc> --step T1 [--surface <path>] [--command <cmd>]`,
		},
		example: "odachi delegate <taskdoc> --step T1",
	},
	{
		name: "level",
		usage: []string{
			`  odachi level [--to <profile>]`,
		},
		example: "odachi level",
	},
	{
		name: "summary",
		usage: []string{
			`  odachi summary`,
		},
		example: "odachi summary",
	},
	{
		name: "design",
		usage: []string{
			`  odachi design <list|comments|open|stop|handoff|sync>`,
		},
		example: "odachi design list | odachi design comments <slug>",
	},
	{
		name: "progress",
		usage: []string{
			`  odachi progress`,
		},
		example: "odachi progress",
	},
	{
		name: "undo",
		usage: []string{
			`  odachi undo [--list | --collect] [--file brief] [--artifact <path>]`,
		},
		example: "odachi undo",
	},
	{
		name: "clean",
		usage: []string{
			`  odachi clean [--apply] [--legacy] [--json]`,
		},
		example: "odachi clean",
	},
	{
		name: "store",
		usage: []string{
			`  odachi store move [--dry-run]`,
		},
		example: "odachi store move",
	},
	{
		name: "evidence",
		usage: []string{
			`  odachi evidence export <tree> --out <path>`,
		},
		example: "odachi evidence export",
	},
	{
		name: "check",
		usage: []string{
			`  odachi check [flags]`,
		},
		example: "odachi check",
	},
	{
		name: "review",
		usage: []string{
			`  odachi review [--exec "<command>" | --attach <file> | --show | --refute <verdict>]`,
			`  odachi review [--lens <lens>]`,
			`  odachi review mode [status|on|off]`,
			`  odachi review checks [status|approve]`,
			`  odachi review waive --because "<why>" [--row <fingerprint>]`,
			`  odachi review probe [--runs N] [--blind] [--exec "<command>"]`,
		},
		example: "odachi review",
	},
	{
		name: "verify",
		usage: []string{
			`  odachi verify [--exec "<command>"] [--model <id>]`,
		},
		example: `odachi verify [--exec "<command>"] [--model <id>]`,
	},
	{
		name: "doctor",
		usage: []string{
			`  odachi doctor [--json]`,
		},
		example: "odachi doctor",
	},
	{
		name: "bench",
		usage: []string{
			`  odachi bench [--journeys <names>] [--binary <path>] [--json]`,
		},
		example: "odachi bench",
	},
	{
		name: "init",
		usage: []string{
			`  odachi init [--skills home|project] [--agents opencode|claude|both] [--model <client>=<model>] [--json]`,
		},
		example: "odachi init",
	},
	{
		name: "setup",
		usage: []string{
			`  odachi setup [--agents opencode|claude|both] [--model <client>=<model>] [--dry-run] [--json]`,
		},
		example: "odachi setup",
	},
	{
		name: "tui",
		usage: []string{
			`  odachi tui`,
		},
		example: "odachi tui",
	},
	{
		name: "skills",
		usage: []string{
			`  odachi skills [--check | --write | --resolve <step>]`,
			`  odachi skills eval <skill> [--exec "<command>"] [--runs N] [--timeout N] [--move-home] [--recover]`,
		},
		example: "odachi skills",
	},
	{
		name: "role",
		usage: []string{
			`  odachi role new <slug>`,
			`  odachi role [<slug>]`,
		},
		example: "odachi role new <slug>",
	},
	{
		name: "skill",
		usage: []string{
			`  odachi skill new <slug> --scope project|home`,
		},
		example: "odachi skill new <slug> --scope project|home",
	},
	{
		name: "pack",
		usage: []string{
			`  odachi pack [--agents claude,codex|standard] [--scope project|home] [--dest <folder>] [--dry-run]`,
		},
		example: "odachi pack",
	},
	{
		name: "version",
		usage: []string{
			`  odachi --version`,
			`  odachi version`,
		},
		example: "odachi --version",
	},
}

// availableCommands renders the unknown-command error's list from the catalog:
// one "name (example)" entry per dispatched verb, in catalog order (0360).
func availableCommands() string {
	entries := make([]string, 0, len(verbCatalog))
	for _, v := range verbCatalog {
		entries = append(entries, v.name+" ("+v.example+")")
	}
	return strings.Join(entries, ", ")
}
