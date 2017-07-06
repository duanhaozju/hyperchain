package common

import (
	"path/filepath"
	"os"
	"github.com/urfave/cli"
)

func NewApp(gitCommit, usage string) *cli.App {
	app := cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Author = "Jialei Rong"
	//app.Authors = nil
	app.Email = "garyrong0905@gmail.com"
	if gitCommit != "" {
		app.Version += "-" + gitCommit[:8]
	}
	app.Usage = usage
	return app
}
