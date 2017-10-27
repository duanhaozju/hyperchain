package network

import (
	"fmt"
	"github.com/hyperchain/hyperchain/p2p/message"
	"github.com/hyperchain/hyperchain/p2p/utils"
	"strconv"
	"strings"
	"time"
)

func (hn *HyperNet) Command(args []string, ret *[]string) error {
	if len(args) < 1 {
		*ret = append(*ret, "please specific the network subcommand.")
		return nil
	}

	switch args[0] {
	//list
	// list all host items
	// hostname => ip:addr
	case "list":
		{
			*ret = append(*ret, "list all connections\n")
			logger.Notice("[IPC] network list")
			// review if the map is empty this will be blocked
			if hn.hostClientMap.IsEmpty() {
				*ret = append(*ret, "connection is empty\n")
				return nil
			}
			for t := range hn.hostClientMap.IterBuffered() {
				c := t.Val.(*Client)
				stat := c.stateMachine.Current()
				addr := c.addr
				hostname := c.hostname
				*ret = append(*ret, fmt.Sprintf("hostname: %s addr:%s stat: %s\n", hostname, addr, stat))
			}
		}
	// connect
	// connect to new remote
	case "connect":
		{
			if len(args) < 3 {
				*ret = append(*ret, "invalid connect parameters, format is `network connect [hostname] [ip:port]`")
				break
			}
			hostname := args[1]
			ipaddr := args[2]
			if !strings.Contains(ipaddr, ":") {
				*ret = append(*ret, fmt.Sprintf("%s is not a valid ipaddress, format is ipaddr:port", ipaddr))
				break
			}
			ip := strings.Split(ipaddr, ":")[0]

			if !utils.IPcheck(ip) {
				*ret = append(*ret, fmt.Sprintf("%s is not a valid ipv4 address", ip))
				break
			}

			port_s := strings.Split(ipaddr, ":")[1]
			port, err := strconv.Atoi(port_s)
			if err != nil {
				*ret = append(*ret, fmt.Sprintf("%s valid port", port_s))
				break
			}
			*ret = append(*ret, fmt.Sprintf("connect to a new host: %s =>> %s:%d\n", hostname, ip, port))
			// real connection part
			//add dns item
			err = hn.dns.AddItem(hostname, ipaddr, false)
			if err != nil {
				*ret = append(*ret, fmt.Sprintf("connect to %s failed, reason: %s", hostname, err.Error()))
				break
			}
			// important after add a dns item, it should persist
			err = hn.dns.Persisit()
			if err != nil {
				*ret = append(*ret, fmt.Sprintf("filed to persist hosts file, reason: %s", err.Error()))
				break
			} else {
				*ret = append(*ret, fmt.Sprintf(" success to persist hosts file! for host %s \n", hostname))
			}

			err = hn.Connect(hostname)
			if err != nil {
				*ret = append(*ret, fmt.Sprintf("connect to %s failed, reason: %s", hostname, err.Error()))
				break
			}

			*ret = append(*ret, fmt.Sprintf("connect to %s successful.", hostname))

		}
	// close
	// close the exists remote peer
	// hypernet close
	case "close":
		{
			if len(args) < 2 {
				*ret = append(*ret, "invalid connect parameters, format is `network close [hostname]`")
				break
			}
			hostname := args[1]
			*ret = append(*ret, fmt.Sprintf("now closing the %s 's connection\n", hostname))

			err := hn.DisConnect(hostname)

			if err != nil {
				*ret = append(*ret, fmt.Sprintf("closing the %s 's connection failed. reason: %s", hostname, err.Error()))
				break
			}
			err = hn.dns.Persisit()

			if err != nil {
				*ret = append(*ret, "persist %s 's connection failed, reason: %s", hostname, err.Error())
				break
			}

			*ret = append(*ret, "close the host connection successed.")
		}
	// reconnect
	// reconnect to already exist peer
	case "reconnect":
		{
			*ret = append(*ret, "reconnect to new host")
			if len(args) < 2 {
				*ret = append(*ret, "invalid connect parameters, format is `network reconnect [hostname]`")
				break
			}
			hostname := strings.TrimSpace(args[1])
			*ret = append(*ret, fmt.Sprintf("now reconnecting the %s 's connection\n", hostname))

			ipaddr, err := hn.dns.GetDNS(hostname)
			if err != nil {
				*ret = append(*ret, fmt.Sprintf("reconnecting the %s 's connection failed (g), reason %s\n", hostname, err.Error()))
				break
			}
			err = hn.DisConnect(hostname)
			if err != nil {
				*ret = append(*ret, fmt.Sprintf("reconnecting the %s 's connection failed (d), reason %s\n", hostname, err.Error()))
				break
			}
			err = hn.ConnectByAddr(hostname, ipaddr)
			if err != nil {
				*ret = append(*ret, fmt.Sprintf("reconnecting the %s 's connection failed (c), reason %s\n", hostname, err.Error()))
				break
			}
			*ret = append(*ret, fmt.Sprintf("reconnecting to %s => %s successful\n", hostname, ipaddr))

		}
	case "ping":
		{
			*ret = append(*ret, "ping to host...")
			if len(args) < 2 {
				*ret = append(*ret, "invalid ping parameters, format is `network ping [hostname]`")
				break
			}
			hostname := strings.TrimSpace(args[1])
			*ret = append(*ret, fmt.Sprintf("now ping to %s\n", hostname))

			c, ok := hn.hostClientMap.Get(hostname)
			if !ok {
				*ret = append(*ret, fmt.Sprintf("invalid hostname (%s), the host not connected.\n", hostname))
				break
			}
			in := message.NewPkg([]byte("ping"), message.ControlType_Notify)
			in.SrcHost = hn.server.selfIdentifier
			start := time.Now().UnixNano()
			_, err := c.(*Client).Discuss(in)
			if err != nil {
				*ret = append(*ret, fmt.Sprintf("ping to %s failed, error info %s\n", hostname, err.Error()))
				break
			}
			end := time.Now().UnixNano()
			*ret = append(*ret, fmt.Sprintf("ping to %s successful, latency is %d ns", hostname, (end-start)))

		}
	default:
		*ret = append(*ret, fmt.Sprintf("unsupport subcommand `network %s`", args[0]))
	}
	return nil
}
