package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cli/cli/pkg/cmd/root"
	"github.com/cli/cli/pkg/cmdutil"
	"github.com/cli/cli/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/pflag"
)

func main() {
	var flagError pflag.ErrorHandling
	docCmd := pflag.NewFlagSet("", flagError)
	manPage := docCmd.BoolP("man-page", "", false, "Generate manual pages")
	website := docCmd.BoolP("website", "", false, "Generate website pages")
	dir := docCmd.StringP("doc-path", "", "", "Path directory where you want generate doc files")
	help := docCmd.BoolP("help", "h", false, "Help about any command")

	if err := docCmd.Parse(os.Args); err != nil {
		os.Exit(1)
	}

	if *help {
		_, err := fmt.Fprintf(os.Stderr, "Usage of %s:\n\n%s", os.Args[0], docCmd.FlagUsages())
		if err != nil {
			fatal(err)
		}
		os.Exit(1)
	}

	if *dir == "" {
		fatal("no dir set")
	}

	io, _, _, _ := iostreams.Test()
	rootCmd := root.NewCmdRoot(&cmdutil.Factory{IOStreams: io}, "", "")

	err := os.MkdirAll(*dir, 0755)
	if err != nil {
		fatal(err)
	}

	if *website {
		err = doc.GenMarkdownTreeCustom(rootCmd, *dir, filePrepender, linkHandler)
		if err != nil {
			fatal(err)
		}

		err = genHelpMarkdown(rootCmd, *dir)
		if err != nil {
			fatal(err)
		}
	}

	if *manPage {
		header := &doc.GenManHeader{
			Title:   "gh",
			Section: "1",
			Source:  "", //source and manual are just put at the top of the manpage, before name
			Manual:  "", //if source is an empty string, it's set to "Auto generated by spf13/cobra"
		}
		err = doc.GenManTree(rootCmd, header, *dir)
		if err != nil {
			fatal(err)
		}
	}
}

func filePrepender(filename string) string {
	return `---
layout: manual
permalink: /:path/:basename
---

`
}

func linkHandler(name string) string {
	return fmt.Sprintf("./%s", strings.TrimSuffix(name, ".md"))
}

func fatal(msg interface{}) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func genHelpMarkdown(cmd *cobra.Command, dir string) error {
	var baseHelpCmd *cobra.Command
	commands := cmd.Commands()
	for _, command := range commands {
		if command.Name() == "help" {
			baseHelpCmd = command
			break
		}
	}
	environmentCmd := root.NewHelpTopic("environment")
	helpCommands := []*cobra.Command{baseHelpCmd, environmentCmd}
	for _, command := range helpCommands {
		filename := "gh_help.md"
		if command.Name() != "help" {
			filename = "gh_help_" + command.Name() + ".md"
		}
		path := filepath.Join(dir, filename)
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := f.WriteString(filePrepender(filename)); err != nil {
			return err
		}
		err = doc.GenMarkdownCustom(command, f, linkHandler)
		if err != nil {
			return err
		}
	}
	return nil
}
