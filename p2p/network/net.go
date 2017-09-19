package network

import (
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

var errInvalidRoute = errors.New("route item is invalid, please check it.")

func ParseRoute(dnsItem string) (hostname string, port int, err error) {
	if !strings.Contains(dnsItem, ":") {
		return "", 0, errInvalidRoute
	}
	route_s := strings.Split(dnsItem, ":")
	hostname = route_s[0]
	port, err = strconv.Atoi(route_s[1])
	if err != nil {
		return "", 0, err
	}
	return
}
