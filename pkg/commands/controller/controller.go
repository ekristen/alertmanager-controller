package commands

import (
	"time"

	"github.com/ekristen/alertmanager-controller/pkg/commands/global"
	"github.com/ekristen/alertmanager-controller/pkg/common"
	"github.com/ekristen/alertmanager-controller/pkg/controller"
	"github.com/urfave/cli/v2"
)

type command struct {
}

func (w *command) Execute(c *cli.Context) error {
	opts := &controller.ControllerOpts{
		GCExpired:      c.Bool("gc-expired"),
		GCExpiredDelay: c.Duration("gc-expired-delay"),
	}
	control, err := controller.New(opts)
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

	flags := []cli.Flag{
		&cli.BoolFlag{
			Name:  "gc-expired",
			Usage: "Garabage Collect Expired Silences",
			Value: true,
		},
		&cli.DurationFlag{
			Name:  "gc-expired-delay",
			Usage: "Delay after Expired before Garbage Collecting Silence",
			Value: 5 * time.Minute,
		},
	}

	cliCmd := &cli.Command{
		Name:   "controller",
		Usage:  "controller",
		Action: cmd.Execute,
		Flags:  append(flags, global.Flags()...),
		Before: global.Before,
	}

	common.RegisterCommand(cliCmd)
}
