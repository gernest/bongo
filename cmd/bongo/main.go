package main

import (
	"log"
	"os"

	"github.com/gernest/bongo"

	"github.com/codegangsta/cli"
)

var (
	authors = []cli.Author{
		{"Geofrey Ernest", "geofreyernest@live.com"},
	}
	sourceFlagName = "source"
	appName        = "bongo"
	version        = "0.1.0"
)

func buildFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:   sourceFlagName,
			Usage:  "sets the path to the project soucce files",
			EnvVar: "PROJECT_SOURCE",
		},
	}
}

func build(ctx *cli.Context) {
	wd, _ := os.Getwd()
	src := wd
	if f := ctx.String(sourceFlagName); f != "" {
		src = f
	}
	app := bongo.New()
	err := app.Run(src)
	if err != nil {
		log.Println(err)
	}
}

func main() {
	app := cli.NewApp()
	app.Name = appName
	app.Usage = "Eleant static website generator"
	app.Authors = authors
	app.Version = version
	app.Commands = []cli.Command{
		cli.Command{
			Name:        "build",
			ShortName:   "b",
			Usage:       "build site",
			Description: "build site",
			Action:      build,
			Flags:       buildFlags(),
		},
	}
	app.RunAndExitOnError()
}
