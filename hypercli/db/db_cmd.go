package db

import (
	"github.com/op/go-logging"
	"github.com/urfave/cli"
	"os"
)

var logger *logging.Logger

func init() {
	logger = logging.MustGetLogger("hypercli/db")
}


//NewServerCMD new server related commands.
func NewDBCMD() []cli.Command {
	return []cli.Command{
		{
			Name:   "recover",
			Usage:  "recover db from a dump file",
			Action: recover,
			Flags:  []cli.Flag{
				cli.StringFlag{
					Name:  "path, p",
					Value: "",
					Usage: "specify dump file path",
				},
			},
		},
		{
			Name:   "dump",
			Usage:  "dump db content with a specify db file",
			Action: dump,
		},
	}
}

func recover(c *cli.Context) error {
	path := c.String("path")
	if path == "" {
		logger.Notice("empty file path")
		os.Exit(1)
	}
	return nil
}

func dump(c *cli.Context) error {
	return nil
}
