package main

import (
	"os"
	"path"

	"github.com/rancher/wrangler/pkg/signals"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/ekristen/prom-am-operator/pkg/common"

	_ "github.com/ekristen/prom-am-operator/pkg/commands/controller"
	_ "github.com/ekristen/prom-am-operator/pkg/commands/version"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			// log panics forces exit
			if _, ok := r.(*logrus.Entry); ok {
				os.Exit(1)
			}
			panic(r)
		}
	}()

	ctx := signals.SetupSignalContext()

	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Usage = "prometheus alertmanager operator"
	app.Version = common.AppVersion.Summary
	app.Authors = []*cli.Author{
		{
			Name:  "Erik Kristensen",
			Email: "erik@erikkristensen.com",
		},
	}

	app.Commands = common.GetCommands()
	app.CommandNotFound = func(context *cli.Context, command string) {
		logrus.Fatalf("Command %s not found.", command)
	}

	if err := app.RunContext(ctx, os.Args); err != nil {
		logrus.Fatal(err)
	}
}
