package version

import (
	"fmt"

	"github.com/ekristen/prom-am-operator/pkg/common"
	"github.com/ekristen/prom-am-operator/pkg/version"
	"github.com/urfave/cli/v2"
)

type versionCommand struct {
}

func (w *versionCommand) Execute(c *cli.Context) error {
	fmt.Printf("%s\n", common.AppVersion.Summary)

	fmt.Println(version.Version("v0.0.0"))

	return nil
}

func init() {
	cmd := versionCommand{}

	cliCmd := &cli.Command{
		Name:   "version",
		Usage:  "print version",
		Action: cmd.Execute,
	}

	common.RegisterCommand(cliCmd)
}
