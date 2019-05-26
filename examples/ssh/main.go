package main

import (
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/ssh"
)

const (
	host = "127.0.0.1"
)

func main() {
	config := &ssh.ClientConfig{
		User:            "abser",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.Password("119216"),
		},
	}

	conn, err := ssh.Dial("tcp", host+":22", config)
	if err != nil {
		log.Fatal("Connection with error: ", err)
	}

	session, err := conn.NewSession()
	if err != nil {
		log.Fatal("Session error: ", err)
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err = session.RequestPty("xterm", 80, 40, modes); err != nil {
		log.Fatal("RequestPty error: ", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		log.Fatal("StdoutPipe() error: ", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		log.Fatal("StdinPipe() error: ", err)
	}

	go func() {
		buf := make([]byte, 1024)

		for {
			if nc, err := stdout.Read(buf); err != nil {
				log.Fatal("Read response failed: ", err)
			} else {
				fmt.Print(string(buf[:nc]))
			}

			buf = buf[0:]
		}
	}()

	if err = session.Shell(); err != nil {
		log.Fatal("Shell() error: ", err)
	}

	cmds := []string{
		"ls -la",
		"ip a",
	}

	for _, cmd := range cmds {
		stdin.Write([]byte(cmd + "\n"))

		log.Print("Send command ", cmd)
	}

	time.Sleep(10 * time.Second)
}
