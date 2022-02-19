package command

import (
	"github.com/mitchellh/cli"
	"github.com/spf13/afero"
)

type Meta struct {
	// UI is a user interface representing input and output.
	UI cli.Ui

	// Fs is an afero filesystem.
	Fs afero.Fs
}
