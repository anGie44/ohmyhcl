package command

import (
	"fmt"
	"log"
	"strings"

	"github.com/anGie44/ohmyhcl/tfrefactor/tfrefactor"
	flag "github.com/spf13/pflag"
)

type ResourceCommand struct {
	Meta
	typ                 string
	providerVersion     string
	path                string
	csv                 bool
	recursive           bool
	ignoreArguments     []string
	ignoreResourceNames []string
	ignorePaths         []string
}

func (r *ResourceCommand) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("resource", flag.ContinueOnError)
	cmdFlags.StringVarP(&r.providerVersion, "provider-version", "p", "latest", "A new provider version constraint")
	cmdFlags.BoolVarP(&r.csv, "csv", "c", false, "Generate .csv file with list of new resources and their parent resource")
	cmdFlags.BoolVarP(&r.recursive, "recursive", "r", false, "Check a directory recursively")
	cmdFlags.StringSliceVarP(&r.ignoreArguments, "ignore-arguments", "", []string{}, "Arguments to ignore")
	cmdFlags.StringSliceVarP(&r.ignoreResourceNames, "ignore-names", "", []string{}, "Specific resource names to ignore")
	cmdFlags.StringSliceVarP(&r.ignorePaths, "ignore-paths", "i", []string{}, "A regular expression for paths to ignore")

	if err := cmdFlags.Parse(args); err != nil {
		r.UI.Error(fmt.Sprintf("failed to parse CLI arguments: %s", err))
		return 1
	}

	if len(cmdFlags.Args()) != 2 { //nolint:gomnd
		r.UI.Error(fmt.Sprintf("The command expects 2 arguments, but got %d", len(cmdFlags.Args())))
		r.UI.Error(r.Help())
		return 1
	}

	r.typ = cmdFlags.Arg(0)
	r.path = cmdFlags.Arg(1)

	log.Printf("[INFO] Migrate resources of type %s to provider version %s", r.typ, r.providerVersion)
	option, err := tfrefactor.NewOption("resource", r.typ, r.providerVersion, r.csv, r.recursive, r.ignoreArguments, r.ignoreResourceNames, r.ignorePaths)
	if err != nil {
		r.UI.Error(err.Error())
		return 1
	}

	log.Printf("[INFO] Migrating file or dir at path: %s", r.path)

	err = tfrefactor.MigrateFileOrDir(r.Fs, r.path, option)
	if err != nil {
		r.UI.Error(err.Error())
		return 1
	}

	return 0
}

// Help returns long-form help text.
func (r *ResourceCommand) Help() string {
	helpText := `
Usage: tfrefactor resource <RESOURCE_TYPE> <PATH> [options]
Arguments
  RESOURCE_TYPE      The provider resource type (e.g. aws_s3_bucket)
  PATH               A path of file or directory to update
Options:
  --ignore-arguments       The arguments in the <RESOURCE_TYPE> to ignore
                           Set the flag with values separated by commas (e.g. --ignore-arguments="acl,grant") or set the flag multiple times.
  --ignore-names           The resource names of <RESOURCE_TYPE> to ignore
                           Set the flag with values separated by commas (e.g. --ignore-names="example,log_bucket") or set the flag multiple times.
  -i  --ignore-paths       Regular expressions for path to ignore
                           Set the flag with values separated by commas or set the flag multiple times.
  -c  --csv    			   Generate a CSV file of new resources and their parent resource (default: false)           
  -p  --provider-version   The provider version constraint (default: v4.0.0)
  -r  --recursive          Check a directory recursively (default: false)
`
	return strings.TrimSpace(helpText)
}

// Synopsis returns one-line help text.
func (r *ResourceCommand) Synopsis() string {
	return "Refactor resource arguments to individual resources"
}
