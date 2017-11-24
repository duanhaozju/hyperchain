package api
//
//import (
//	"fmt"
//	"strconv"
//)
//
//const (
//	START   = "start"
//	STOP    = "stop"
//	RESTART = "restart"
//	HTTP    = "http"
//	WS      = "websocket"
//)
//
//func (asi *APIServerImpl) Command(args []string, ret *[]string) error {
//	if len(args) < 1 {
//		*ret = append(*ret, "please specific the service subcommand.")
//		return nil
//	}
//
//	switch args[0] {
//	// http
//	// start, stop or restart http service
//	case HTTP:
//		{
//			if len(args) < 2 {
//				*ret = append(*ret, "invalid parameters, please input your operation: start, stop or restart")
//				break
//			}
//			op := args[1]
//
//			switch op {
//			case START:
//				{
//					if len(args) < 3 {
//						*ret = append(*ret, "invalid parameters, format is `service http start [port]`")
//						break
//					}
//
//					port := args[2]
//					i, err := strconv.ParseInt(port, 10, 0)
//					if err != nil {
//						*ret = append(*ret, "invalid port, please give decimal integer")
//						break
//					}
//
//					log.Noticef("[IPC] service http %v %v", args[1], args[2])
//					asi.httpServer.setPort(int(i))
//					if err := asi.httpServer.start(); err != nil {
//						*ret = append(*ret, fmt.Sprintf("failed to start http service at port %s. Err: %v", port, err))
//					} else {
//						*ret = append(*ret, fmt.Sprintf("success to start http service at port %s.", port))
//					}
//					break
//				}
//			case STOP:
//				{
//					log.Noticef("[IPC] service http %v", args[1])
//					*ret = append(*ret, fmt.Sprintf("waitting for %v to stop...\n", ReadTimeout))
//					if err := asi.httpServer.stop(); err != nil {
//						*ret = append(*ret, fmt.Sprintf("failed to stop http service. Err: %v", err))
//					} else {
//						*ret = append(*ret, fmt.Sprintf("success to stop http service at port %v", asi.httpServer.getPort()))
//					}
//					break
//				}
//			case RESTART:
//				{
//					log.Noticef("[IPC] service http %v", args[1])
//					if err := asi.httpServer.restart(); err != nil {
//						*ret = append(*ret, fmt.Sprintf("failed to restart http service. Err: %v", err))
//					} else {
//						*ret = append(*ret, "success to restart http service.")
//					}
//					break
//				}
//			default:
//				*ret = append(*ret, fmt.Sprintf("unknown command `%v`, please input your operation: start, stop or restart.", op))
//			}
//		}
//
//	// websocket
//	// start, stop or restart websocket service
//	case WS:
//		{
//			if len(args) < 2 {
//				*ret = append(*ret, "invalid parameters, please input your operation: start, stop or restart")
//				break
//			}
//			op := args[1]
//
//			switch op {
//			case START:
//				{
//					if len(args) < 3 {
//						*ret = append(*ret, "invalid parameters, format is `service websocket start [port]`")
//						break
//					}
//
//					port := args[2]
//					i, err := strconv.ParseInt(port, 10, 0)
//					if err != nil {
//						*ret = append(*ret, "invalid port, please give decimal integer")
//						break
//					}
//
//					log.Noticef("[IPC] service websocket %v %v", args[1], args[2])
//					asi.wsServer.setPort(int(i))
//					if err := asi.wsServer.start(); err != nil {
//						*ret = append(*ret, fmt.Sprintf("failed to start websocket service at port %s. Err: %v", port, err))
//					} else {
//						*ret = append(*ret, fmt.Sprintf("success to start websocket service at port %s.", port))
//					}
//					break
//				}
//			case STOP:
//				{
//					log.Noticef("[IPC] service websocket %v", args[1])
//					if err := asi.wsServer.stop(); err != nil {
//						*ret = append(*ret, fmt.Sprintf("failed to stop websocket service. Err: %v", err))
//					} else {
//						*ret = append(*ret, fmt.Sprintf("success to stop websocket service at port %v", asi.wsServer.getPort()))
//					}
//					break
//				}
//			case RESTART:
//				{
//					log.Noticef("[IPC] service websocket %v", args[1])
//					if err := asi.wsServer.restart(); err != nil {
//						*ret = append(*ret, fmt.Sprintf("failed to restart websocket service. Err: %v", err))
//					} else {
//						*ret = append(*ret, "success to restart websocket service.")
//					}
//					break
//				}
//			default:
//				*ret = append(*ret, fmt.Sprintf("unknown command `%v`, please input your operation: start, stop or restart.", op))
//			}
//		}
//	default:
//		*ret = append(*ret, fmt.Sprintf("unsupport subcommand `service %s`", args[0]))
//	}
//	return nil
//}
