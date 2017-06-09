package network

import (
	"strings"
	"github.com/pkg/errors"
	"strconv"
)
var errInvalidRoute = errors.New("route item is invalid, please check it.")

func ParseRoute(routeItem string) (hostname string,port int,err error){
	if !strings.Contains(routeItem,":"){
		return "",0,errInvalidRoute
	}
	route_s := strings.Split(routeItem,":")
	hostname = route_s[0]
	port,err = strconv.Atoi(route_s[1])
	if err != nil{
		return "",0,err
	}
	return
}
