package clis

import (
	"github.com/moqsien/goutils/pkgs/gtea/gprint"
	"github.com/spf13/cobra"
)

type Cli struct {
	rootCmd *cobra.Command
	groupID string
	gitTag  string
	gitHash string
	gitTime string
}

const (
	GroupID = "gvc"
)

func New() (c *Cli) {
	c = &Cli{
		rootCmd: &cobra.Command{
			Short: "Geek's Versatile Crafts",
			Long:  "g <Command> <SubCommand> --flags args...",
		},
		groupID: GroupID,
	}
	c.rootCmd.AddGroup(&cobra.Group{ID: c.groupID, Title: "Command list: "})
	c.initiate()
	return
}

func (that *Cli) initiate() {
	// self related CLIs
	that.showVersion()
	that.checkForUpdate()
	that.uninstall()
	that.configure()
	that.ssh()
}

func (that *Cli) Run() {
	if err := that.rootCmd.Execute(); err != nil {
		gprint.PrintError("%+v", err)
	}
}
