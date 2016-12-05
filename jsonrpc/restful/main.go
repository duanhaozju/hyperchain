package restful

import (
	"github.com/astaxie/beego"
	_ "hyperchain/jsonrpc/restful/routers"
)

func main() {
	beego.Run("127.0.0.1:9000")
}
