package acli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CommandInfo describes a command in a machine-readable format for agent discovery.
type CommandInfo struct {
	Name        string        `json:"name"`
	FullCommand string        `json:"full_command"`
	Description string        `json:"description"`
	Aliases     []string      `json:"aliases,omitempty"`
	Usage       string        `json:"usage"`
	Args        string        `json:"args,omitempty"`
	Flags       []FlagInfo    `json:"flags,omitempty"`
	Subcommands []CommandInfo `json:"subcommands,omitempty"`
}

// FlagInfo describes a flag in a machine-readable format.
type FlagInfo struct {
	Name      string `json:"name"`
	Shorthand string `json:"shorthand,omitempty"`
	Type      string `json:"type"`
	Default   string `json:"default,omitempty"`
	Required  bool   `json:"required,omitempty"`
	Usage     string `json:"usage"`
}

var commandsCmd = &cobra.Command{
	Use:   "commands",
	Short: "List all commands in machine-readable JSON (for agents and scripts)",
	Long: `Outputs the full command tree as JSON, including all subcommands, flags,
aliases, and descriptions. Designed for programmatic consumption by LLM coding
agents and automation scripts that need to discover available operations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tree := buildCommandTree(rootCmd)
		return outputJSON(tree)
	},
}

func buildCommandTree(cmd *cobra.Command) CommandInfo {
	info := CommandInfo{
		Name:        cmd.Name(),
		FullCommand: cmd.CommandPath(),
		Description: cmd.Short,
		Aliases:     cmd.Aliases,
		Usage:       cmd.UseLine(),
	}

	// Determine expected args from Use string
	if cmd.Args != nil {
		info.Args = extractArgsFromUse(cmd.Use)
	}

	// Collect local flags (not inherited)
	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if f.Hidden {
			return
		}
		fi := FlagInfo{
			Name:      f.Name,
			Shorthand: f.Shorthand,
			Type:      f.Value.Type(),
			Default:   f.DefValue,
			Usage:     f.Usage,
		}
		// Check if flag is required via annotations
		if ann, ok := f.Annotations[cobra.BashCompOneRequiredFlag]; ok {
			for _, v := range ann {
				if v == "true" {
					fi.Required = true
				}
			}
		}
		info.Flags = append(info.Flags, fi)
	})

	// Recurse into subcommands
	for _, sub := range cmd.Commands() {
		if sub.Hidden || sub.Name() == "help" || sub.Name() == "completion" {
			continue
		}
		info.Subcommands = append(info.Subcommands, buildCommandTree(sub))
	}

	return info
}

// extractArgsFromUse extracts the argument portion from a cobra Use string.
// e.g. "get <issue-key>" -> "<issue-key>"
func extractArgsFromUse(use string) string {
	for i, c := range use {
		if c == ' ' {
			return use[i+1:]
		}
	}
	return ""
}
