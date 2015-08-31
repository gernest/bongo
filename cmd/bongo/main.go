package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gernest/bongo"

	"gopkg.in/fsnotify.v1"

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

func serve(ctx *cli.Context) {
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
	files, err := bongo.NewLoader().Load(src)
	if err != nil {
		log.Fatal(err)
	}
	watch, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watch.Close()
	for _, file := range files {
		watch.Add(file)
	}
	go func() {
		dir := filepath.Join(src, bongo.OutputDir)
		log.Println("serving website", dir, "  at  http://localhost:8000")
		log.Fatal(http.ListenAndServe(":8000", http.FileServer(http.Dir(dir))))
	}()
	for {
		select {
		case event := <-watch.Events:
			if event.Op&(fsnotify.Rename|fsnotify.Create|fsnotify.Write) > 0 {
				log.Printf("detected change %s  Rebuilding...\n", event.Name)
				app.Run(src)
			}
		case err := <-watch.Errors:
			if err != nil {
				log.Println(err)
			}
		default:
			continue
		}
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
		cli.Command{
			Name:        "serve",
			ShortName:   "s",
			Usage:       "builds and serves the project",
			Description: "serves site",
			Action:      serve,
			Flags:       buildFlags(),
		},
	}
	app.RunAndExitOnError()
}
