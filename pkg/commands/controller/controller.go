package commands

import (
	"github.com/ekristen/alertmanager-controller/pkg/commands/global"
	"github.com/ekristen/alertmanager-controller/pkg/common"
	"github.com/ekristen/alertmanager-controller/pkg/controller"
	"github.com/urfave/cli/v2"
)

type command struct {
}

func (w *command) Execute(c *cli.Context) error {
	control, err := controller.New()
	if err != nil {
		return err
	}
	if err := control.Start(c.Context); err != nil {
		return err
	}
	<-c.Context.Done()
	return nil
}

func init() {
	cmd := command{}

	cliCmd := &cli.Command{
		Name:   "controller",
		Usage:  "controller",
		Action: cmd.Execute,
		Flags:  global.Flags(),
	}

	common.RegisterCommand(cliCmd)
}
