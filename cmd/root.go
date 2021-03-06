package cmd

import (
	"fmt"
	"log"
	"os"
	"runtime"

	clilog "github.com/b4b4r07/go-cli-log"
	"github.com/spf13/cobra"
)

var (
	// Version is the version number
	Version = "unset"

	// BuildTag set during build to git tag, if any
	BuildTag = "unset"

	// BuildSHA is the git sha set during build
	BuildSHA = "unset"
)

// newRootCmd returns the root command
func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:                "blog",
		Short:              "A CLI tool for editing blog built by hugo etc",
		SilenceErrors:      true,
		DisableSuggestions: false,
		Version:            fmt.Sprintf("%s (%s/%s)", Version, BuildTag, BuildSHA),
	}

	rootCmd.AddCommand(newEditCmd())
	rootCmd.AddCommand(newNewCmd())
	return rootCmd
}

// Execute is
func Execute() error {
	clilog.Env = "BLOG_LOG"
	clilog.Path = "BLOG_LOG_PATH"
	clilog.SetOutput()

	log.Printf("[INFO] pkg version: %s", Version)
	log.Printf("[INFO] Go runtime version: %s", runtime.Version())
	log.Printf("[INFO] Build tag/SHA: %s/%s", BuildTag, BuildSHA)
	log.Printf("[INFO] CLI args: %#v", os.Args)

	defer log.Printf("[DEBUG] root command execution finished")

	return newRootCmd().Execute()
}
