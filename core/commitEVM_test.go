package core

import "testing"

func TestCommitStatedbToBlockchain(t *testing.T){
	GetVMEnv().State().Commit()
}