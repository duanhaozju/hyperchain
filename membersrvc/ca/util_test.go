//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package ca

import (
	"reflect"
	"testing"

	pb "hyperchain/membersrvc/protos"
)

func TestRandomString(t *testing.T) {

	rand := randomString(100)
	if rand == "" {
		t.Error("The randomString function must return a non nil value.")
	}

	if reflect.TypeOf(rand).String() != "string" {
		t.Error("The randomString function must returns a string type value.")
	}
}

func TestMemberRoleToStringNone(t *testing.T) {

	var role, err = MemberRoleToString(pb.Role_NONE)

	if err != nil {
		t.Errorf("Error converting member role to string, result: %s", role)
	}

	if role != pb.Role_name[int32(pb.Role_NONE)] {
		t.Errorf("The role string returned should have been %s", "NONE")
	}

}
