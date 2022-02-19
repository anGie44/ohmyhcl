package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/anGie44/ohmyhcl/tfrefactor/command"
	"github.com/hashicorp/logutils"
	"github.com/mitchellh/cli"
	"github.com/spf13/afero"
)

var UI cli.Ui

func init() {
	UI = &cli.BasicUi{
		Writer: os.Stdout,
	}
}

func main() {
	log.SetOutput(logOutput())
	log.Printf("[INFO] CLI args: %#v", os.Args)

	commands := initCommands()

	args := os.Args[1:]

	c := &cli.CLI{
		Name:                  "tfrefactor",
		Args:                  args,
		Commands:              commands,
		HelpWriter:            os.Stdout,
		Autocomplete:          true,
		AutocompleteInstall:   "install-autocomplete",
		AutocompleteUninstall: "uninstall-autocomplete",
	}

	exitStatus, err := c.Run()
	if err != nil {
		UI.Error(fmt.Sprintf("Failed to execute CLI: %s", err))
	}

	os.Exit(exitStatus)
}

func logOutput() io.Writer {
	levels := []logutils.LogLevel{"TRACE", "DEBUG", "INFO", "WARN", "ERROR"}
	minLevel := os.Getenv("TFREFACTOR_LOG")

	// default log writer is null device.
	writer := ioutil.Discard
	if minLevel != "" {
		writer = os.Stderr
	}

	filter := &logutils.LevelFilter{
		Levels:   levels,
		MinLevel: logutils.LogLevel(strings.ToUpper(minLevel)),
		Writer:   writer,
	}

	return filter
}

func initCommands() map[string]cli.CommandFactory {
	meta := command.Meta{
		UI: UI,
		Fs: afero.NewOsFs(),
	}

	commands := map[string]cli.CommandFactory{
		"resource": func() (cli.Command, error) {
			return &command.ResourceCommand{
				Meta: meta,
			}, nil
		},
	}

	return commands
}
