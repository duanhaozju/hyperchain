package utils

import "testing"

func TestGetProjectPath(t *testing.T) {
	t.Log(GetProjectPath())
	t.Log(GetCurrentDirectory())
}
