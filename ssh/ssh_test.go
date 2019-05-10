package moressh

import "testing"

var (
	user     = "root"
	password = ""
	host     = ""
	port     = 22
	cmds     = []string{
		"ls -la",
		"ip a",
	}
	timeout = 10
)

func Test_Do(t *testing.T) {
	chs := make(chan Result, 1)
	Do(user, password, host, cmds, port, timeout, chs)
}
