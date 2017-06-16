//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package jcee

import (
	"github.com/op/go-logging"
	"hyperchain/common"
	"os/exec"
	//"strconv"
)

//HyperVM general remote VM interface
type HyperVM interface {
	Start() error
	Stop() error
}

//HyperJVM an implementation of HyperVM
type HyperJVM struct {
	ledgerPort int
	jceePort   int // java contract execution engine
	startCmd   *exec.Cmd
	logger     *logging.Logger
}

//Start start the hyerjvm
func (hjvm *HyperJVM) Start() error {
	//exec.Command("cd ./hyperjvm/bin").Run()
	//hjvm.logger.Critical("try to start jvm")
	//cmd := exec.Command("./hyperjvm/bin/start_hyperjvm.sh", strconv.Itoa(hjvm.jceePort), strconv.Itoa(hjvm.ledgerPort))
	//
	//hjvm.startCmd = cmd
	//err := hjvm.startCmd.Run()
	//output, _ := cmd.Output()
	//hjvm.logger.Error(string(output))
	//if err != nil {
	//	hjvm.logger.Error(err)
	//	return err
	//}
	//
	//hjvm.logger.Info("execute start hyperjvm command successful")
	//exec.Command("cd ../../../").Run()
	return nil
}

//Stop stop hyperjvm
func (hjvm *HyperJVM) Stop() error {
	if hjvm.startCmd != nil {
		err := hjvm.startCmd.Process.Kill()
		if err != nil {
			hjvm.logger.Error(err)
			return err
		}
		hjvm.logger.Info("kill hjvm process")
	}
	return nil
}

func NewHyperJVM(lport, jport int) HyperVM {
	hj := &HyperJVM{ledgerPort: lport, jceePort: jport}
	hj.logger = common.GetLogger(common.DEFAULT_LOG, "jcee")
	return hj
}
