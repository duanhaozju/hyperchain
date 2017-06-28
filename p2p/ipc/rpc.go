package ipc

import (
	"sync"
	"fmt"
	"errors"
)

type Args struct {
	Cmd string
	Argv []string
}

type Ret struct {
	Rets []string
}

type RemoteCall struct{
	RegLock *sync.RWMutex
	Func map[string]func(args []string,ret *[]string)error
}

func NewRemoteCall()*RemoteCall{
	return &RemoteCall{
		RegLock:new(sync.RWMutex),
		Func:make(map[string]func(args []string,ret *[]string)error),
	}
}

func RegisterFunc(rc *RemoteCall,cmd string,f func(args []string,ret *[]string)error ){
	rc.RegLock.Lock()
	defer rc.RegLock.Unlock()
	rc.Func[cmd] = f
}

func(rc *RemoteCall)Call(args Args,ret *Ret)error{
	rc.RegLock.RLock()
	defer rc.RegLock.RUnlock()
	if f,ok := rc.Func[args.Cmd];ok{
		return f(args.Argv,&ret.Rets)
	}else{
		return errors.New(fmt.Sprintf("the cmd: %s hasn't register.",args.Cmd))
	}
}

