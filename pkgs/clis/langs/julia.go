package langs

import (
	"github.com/moqsien/gvc/pkgs/vctrl"
	"github.com/spf13/cobra"
)

func SetJulia(reg IRegister) {
	nodeCmd := &cobra.Command{
		Use:     "julia",
		Aliases: []string{"jl", "J"},
		Short:   "Julia related CLIs.",
	}

	remoteCmd := &cobra.Command{
		Use:     "remote",
		Aliases: []string{"r"},
		Short:   "Shows available versions from remote website.",
		Run: func(cmd *cobra.Command, args []string) {
			jv := vctrl.NewJuliaVersion()
			jv.ShowVersions()
		},
	}
	nodeCmd.AddCommand(remoteCmd)

	useCmd := &cobra.Command{
		Use:     "use",
		Aliases: []string{"u"},
		Short:   "Downloads and switches to the specified version.",
		Long:    "Example: J u <version>",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				return
			}
			jv := vctrl.NewJuliaVersion()
			jv.UseVersion(args[0])
		},
	}
	nodeCmd.AddCommand(useCmd)

	localCmd := &cobra.Command{
		Use:     "local",
		Aliases: []string{"l"},
		Short:   "Shows installed versions.",
		Run: func(cmd *cobra.Command, args []string) {
			jv := vctrl.NewJuliaVersion()
			jv.ShowInstalled()
		},
	}
	nodeCmd.AddCommand(localCmd)

	removeAllCmd := &cobra.Command{
		Use:     "remove-unused",
		Aliases: []string{"ru"},
		Short:   "Removes installed versions except the one currently in use.",
		Run: func(cmd *cobra.Command, args []string) {
			jv := vctrl.NewJuliaVersion()
			jv.RemoveUnused()
		},
	}
	nodeCmd.AddCommand(removeAllCmd)

	removeCmd := &cobra.Command{
		Use:     "remove",
		Aliases: []string{"rm"},
		Short:   "Removes a specified version.",
		Long:    "Example: J rm <version>",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				return
			}
			jv := vctrl.NewJuliaVersion()
			jv.RemoveVersion(args[0])
		},
	}
	nodeCmd.AddCommand(removeCmd)
	reg.Register(nodeCmd)
}
