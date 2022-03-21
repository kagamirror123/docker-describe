package main

import (
	"github.com/docker/cli/cli-plugins/manager"
	"github.com/docker/cli/cli-plugins/plugin"
	"github.com/docker/cli/cli/command"
	"github.com/kagamirror123/docker-psx/commands"
	"github.com/spf13/cobra"
)

func main() {

	plugin.Run(func(dockerCli command.Cli) *cobra.Command {
		return commands.NewRootCmd("psx", dockerCli)
	},
		manager.Metadata{
			SchemaVersion:    "0.1.0",
			Vendor:           "nkagamirror",
			Version:          "0.0.1",
			ShortDescription: "this plugin display containers by compose project",
			URL:              "https://github.com/kagamirror123/docker-psx",
		})
}
