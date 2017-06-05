package db_helper
import (
	"github.com/urfave/cli"
	"os"
	"fmt"
	"time"
)

func main() {

	app := cli.NewApp()
	app.Name = "LevelDB Util"
	app.Version = "1.3.0"
	app.Compiled = time.Now()
	app.Usage = "Hyperchain db parser"
	app.Description = "Run 'hypercli COMMAND --help' for more information on a command"
	app.Action = AppEntry
	app.Run(os.Args)
}


func AppEntry(c *cli.Context) error {

}
