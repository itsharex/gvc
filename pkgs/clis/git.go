package clis

import (
	"runtime"

	"github.com/gvcgo/gvc/pkgs/utils"
	"github.com/gvcgo/gvc/pkgs/vctrl"
	"github.com/spf13/cobra"
)

// github download acceleration.
func (that *Cli) github() {
	githubCmd := &cobra.Command{
		Use:     "github",
		Aliases: []string{"gh"},
		Short:   "Github related CLIs.",
		GroupID: that.groupID,
	}

	// automatically modifies hosts file.
	githubCmd.AddCommand(&cobra.Command{
		Use:     "hosts",
		Aliases: []string{"ho"},
		Short:   "Modifies hosts file for github.",
		Run: func(cmd *cobra.Command, args []string) {
			h := vctrl.NewHosts()
			if runtime.GOOS != utils.Windows {
				h.Run()
			} else {
				h.WinRunAsAdmin()
			}
			h.ShowFilePath()
		},
	})

	var (
		destHostName string = "dest_host"
		destPortName string = "dest_port"
		timeoutName  string = "timeout"
	)
	sshProxyHttp := &cobra.Command{
		Use:     "crokscrew",
		Aliases: []string{"cs"},
		Short:   "Http proxy for ssh.",
		Long:    "Example: g gh cs --dest_host=xxx --dest_port=xxx --timeout=xxx",
		Run: func(cmd *cobra.Command, args []string) {
			destHost, _ := cmd.Flags().GetString(destHostName)
			destPort, _ := cmd.Flags().GetString(destPortName)
			timeout, _ := cmd.Flags().GetInt(timeoutName)
			if destHost == "" || destPort == "" {
				cmd.Help()
				return
			}
			vctrl.GrokscrewHttpSSH(destHost, destPort, timeout)
		},
	}

	sshProxyHttp.Flags().StringP(destHostName, "a", "", "Specifies dest host.")
	sshProxyHttp.Flags().StringP(destPortName, "p", "", "Specifies dest port.")
	sshProxyHttp.Flags().IntP(timeoutName, "t", 3, "Specifies timeout.")
	githubCmd.AddCommand(sshProxyHttp)

	// vg := vctrl.NewGhDownloader()

	// githubCmd.AddCommand(&cobra.Command{
	// 	Use:     "proxy",
	// 	Aliases: []string{"p"},
	// 	Short:   "Sets a proxy URI for github downloads.",
	// 	Long:    "Example: gh p https://gh.flyinbug.top/gh/",
	// 	Run: func(cmd *cobra.Command, args []string) {
	// 		if len(args) == 0 {
	// 			cmd.Help()
	// 			return
	// 		}
	// 		vg.SetReverseProxyForDownload(args[0])
	// 	},
	// })

	// sourceCodeFlag := "code"
	// download := &cobra.Command{
	// 	Use:     "download",
	// 	Aliases: []string{"d"},
	// 	Short:   "Downloads released files or source code from a github repo.",
	// 	Long:    "Example: gh d http://github.com/gvcgo/gvc",
	// 	Run: func(cmd *cobra.Command, args []string) {
	// 		if len(args) == 0 {
	// 			cmd.Help()
	// 			return
	// 		}
	// 		onlySourceCode, _ := cmd.Flags().GetBool(sourceCodeFlag)
	// 		vg.Download(args[0], onlySourceCode)
	// 	},
	// }
	// download.Flags().BoolP(sourceCodeFlag, "c", false, "To download source code only.")
	// githubCmd.AddCommand(download)

	that.rootCmd.AddCommand(githubCmd)
}

// git related CLIs.
func (that *Cli) git() {
	gitCmd := &cobra.Command{
		Use:     "git",
		Aliases: []string{"g"},
		Short:   "Git related CLIs(especially with a proxy).",
		GroupID: that.groupID,
	}

	vg := vctrl.NewGhDownloader()

	// .ssh files.
	gitCmd.AddCommand(&cobra.Command{
		Use:     "upload-dotssh",
		Aliases: []string{"u"},
		Short:   "Uploads local .ssh files to remote repo after zipped with password.",
		Run: func(cmd *cobra.Command, args []string) {
			vg.HandleDotSSHFiles(false)
		},
	})

	gitCmd.AddCommand(&cobra.Command{
		Use:     "download-dotssh",
		Aliases: []string{"D"},
		Short:   "Downloads .ssh files from remote repo, then deploy them.",
		Run: func(cmd *cobra.Command, args []string) {
			vg.HandleDotSSHFiles(true)
		},
	})

	// Windows git.exe.
	gitCmd.AddCommand(&cobra.Command{
		Use:     "install",
		Aliases: []string{"i"},
		Short:   "Installs git for windows.",
		Run: func(cmd *cobra.Command, args []string) {
			if runtime.GOOS != utils.Windows {
				cmd.Help()
				return
			}
			vg.InstallGitForWindows()
		},
	})

	// Set local proxy for git.
	gitCmd.AddCommand(&cobra.Command{
		Use:     "proxy",
		Aliases: []string{"pr"},
		Short:   "Sets a proxy for your git.",
		Long:    "Example: g pr http://localhost:2023",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				return
			}
			vg.SaveDefaultProxy(args[0])
		},
	})

	sshProxyCmd := &cobra.Command{
		Use:     "ssh-proxy-fix",
		Aliases: []string{"spf"},
		Short:   "Adds proxy info to the ssh config file.",
		Run: func(cmd *cobra.Command, args []string) {
			pxyURI := vg.ReadDefaultProxy()
			vg.SetProxyForGitSSH(pxyURI)
		},
	}
	gitCmd.AddCommand(sshProxyCmd)

	toggleProxyForGitSSHCmd := &cobra.Command{
		Use:     "toggle-ssh-proxy",
		Aliases: []string{"tsp"},
		Short:   "Toggle status of the proxy for ssh.",
		Run: func(cmd *cobra.Command, args []string) {
			vg.ToggleProxyForGitSSH()
		},
	}
	gitCmd.AddCommand(toggleProxyForGitSSHCmd)

	var toEnableProxyFlagName string = "enable-proxy"
	lazyGitCmd := &cobra.Command{
		Use:     "lazygit",
		Aliases: []string{"lg"},
		Short:   "Start lazygit with/without an ssh proxy.",
		Long:    "Example: g git lg -e [layzgit-args...]",
		Run: func(cmd *cobra.Command, args []string) {
			p, _ := cmd.Flags().GetBool(toEnableProxyFlagName)
			vg.LazyGit(p, args...)
		},
	}
	lazyGitCmd.Flags().BoolP(toEnableProxyFlagName, "e", false, "To enable the proxy for lazygit.")
	gitCmd.AddCommand(lazyGitCmd)

	var (
		defaultProxy        string = vg.ReadDefaultProxy()
		manualProxyFlagName string = "proxy"
		NoProxyFlagName     string = "no"
		workDirName         string = "workdir"
	)

	getProxy := func(cmd *cobra.Command) string {
		pxy := ""
		if disableProxy, _ := cmd.Flags().GetBool(NoProxyFlagName); !disableProxy {
			pxy, _ = cmd.Flags().GetString(manualProxyFlagName)
			if pxy == "" {
				pxy = defaultProxy
			}
		}
		return pxy
	}

	setWorkDir := func(cmd *cobra.Command) {
		wd, _ := cmd.Flags().GetString(workDirName)
		vg.SetWorkDir(wd)
	}

	cloneCmd := &cobra.Command{
		Use:     "clone",
		Aliases: []string{"c"},
		Short:   "Clones a remote repo.",
		Long:    "Example: g c --proxy=http://localhost:2023  git@github.com:moqsien/gvc.git",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				return
			}
			setWorkDir(cmd)
			vg.Clone(args[0], getProxy(cmd))
		},
	}
	cloneCmd.Flags().StringP(manualProxyFlagName, "p", "", "Specifies the proxy for using.")
	cloneCmd.Flags().BoolP(NoProxyFlagName, "n", false, "Disables the proxy.")
	cloneCmd.Flags().StringP(workDirName, "d", "", "Sets work dir for go-git.")
	gitCmd.AddCommand(cloneCmd)

	pullCmd := &cobra.Command{
		Use:     "pull",
		Aliases: []string{"P"},
		Short:   "Pulls from a remote repo.",
		Long:    "Example: g P --proxy=http://localhost:2023",
		Run: func(cmd *cobra.Command, args []string) {
			setWorkDir(cmd)
			vg.Pull(getProxy(cmd))
		},
	}
	pullCmd.Flags().StringP(manualProxyFlagName, "p", "", "Specifies the proxy for using.")
	pullCmd.Flags().BoolP(NoProxyFlagName, "n", false, "Disables the proxy.")
	pullCmd.Flags().StringP(workDirName, "d", "", "Sets work dir for go-git.")
	gitCmd.AddCommand(pullCmd)

	pushCmd := &cobra.Command{
		Use:     "push",
		Aliases: []string{"p"},
		Short:   "Pushes to a remot repo.",
		Long:    "Example: g p --proxy=http://localhost:2023",
		Run: func(cmd *cobra.Command, args []string) {
			setWorkDir(cmd)
			vg.Push(getProxy(cmd))
		},
	}
	pushCmd.Flags().StringP(manualProxyFlagName, "p", "", "Specifies the proxy for using.")
	pushCmd.Flags().BoolP(NoProxyFlagName, "n", false, "Disables the proxy.")
	pushCmd.Flags().StringP(workDirName, "d", "", "Sets work dir for go-git.")
	gitCmd.AddCommand(pushCmd)

	commitPushCmd := &cobra.Command{
		Use:     "commit-push",
		Aliases: []string{"cp"},
		Short:   "Commits and pushes to a remote repo.",
		Long:    "Example: g cp --proxy=http://localhost:2023 <commit msg>",
		Run: func(cmd *cobra.Command, args []string) {
			commitMsg := "update"
			if len(args) > 0 {
				commitMsg = args[0]
			}
			setWorkDir(cmd)
			vg.CommitAndPush(commitMsg, getProxy(cmd))
		},
	}
	commitPushCmd.Flags().StringP(manualProxyFlagName, "p", "", "Specifies the proxy for using.")
	commitPushCmd.Flags().BoolP(NoProxyFlagName, "n", false, "Disables the proxy.")
	commitPushCmd.Flags().StringP(workDirName, "d", "", "Sets work dir for go-git.")
	gitCmd.AddCommand(commitPushCmd)

	latesTagCmd := &cobra.Command{
		Use:     "tag-latest",
		Aliases: []string{"tl", "t"},
		Short:   "Shows the latest tag of a local repo.",
		Run: func(cmd *cobra.Command, args []string) {
			setWorkDir(cmd)
			vg.ShowLatestTag()
		},
	}
	latesTagCmd.Flags().StringP(workDirName, "d", "", "Sets work dir for go-git.")
	gitCmd.AddCommand(latesTagCmd)

	addTagPushCmd := &cobra.Command{
		Use:     "tag-push",
		Aliases: []string{"tp"},
		Short:   "Adds a tag and pushes to a remote repo.",
		Long:    "Example: g tp v0.0.1",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				return
			}
			setWorkDir(cmd)
			vg.AddTagAndPush(args[0], getProxy(cmd))
		},
	}
	addTagPushCmd.Flags().StringP(manualProxyFlagName, "p", "", "Specifies the proxy for using.")
	addTagPushCmd.Flags().BoolP(NoProxyFlagName, "n", false, "Disables the proxy.")
	addTagPushCmd.Flags().StringP(workDirName, "d", "", "Sets work dir for go-git.")
	gitCmd.AddCommand(addTagPushCmd)

	delTagCmd := &cobra.Command{
		Use:     "detag-push",
		Aliases: []string{"dp"},
		Short:   "Deletes a tag and pushes to a remote repo.",
		Long:    "Example: g dp v0.0.1",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				return
			}
			setWorkDir(cmd)
			vg.DelTagAndPush(args[0], getProxy(cmd))
		},
	}
	delTagCmd.Flags().StringP(manualProxyFlagName, "p", "", "Specifies the proxy for using.")
	delTagCmd.Flags().BoolP(NoProxyFlagName, "n", false, "Disables the proxy.")
	delTagCmd.Flags().StringP(workDirName, "d", "", "Sets work dir for go-git.")
	gitCmd.AddCommand(delTagCmd)

	that.rootCmd.AddCommand(gitCmd)
}
