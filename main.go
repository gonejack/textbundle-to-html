package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gonejack/textbundle-to-html/cmd"
	"github.com/spf13/cobra"
)

var (
	verbose = false

	prog = &cobra.Command{
		Use:   "textbundle-to-html *.textbundle",
		Short: "Command line tool for converting textbundles to html.",
		Run: func(c *cobra.Command, args []string) {
			err := run(c, args)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

func init() {
	log.SetOutput(os.Stdout)

	flags := prog.Flags()
	{
		flags.SortFlags = false
		flags.BoolVarP(&verbose, "verbose", "v", false, "verbose")
	}
}

func run(c *cobra.Command, args []string) error {
	exec := cmd.TextBundleToEpub{
		Verbose: verbose,
	}

	if len(args) == 0 || args[0] == "*.textbundle" {
		args, _ = filepath.Glob("*.textbundle")
	}

	return exec.Run(args)
}

func main() {
	_ = prog.Execute()
}
