package moressh

import (
	"bytes"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"
)

//Result to use []chan Result know concurrency
type Result struct {
	Host    string
	Success bool
	Result  string
}

//session return a session between
func session(user, host, password string, port int) (*ssh.Session, error) {
	config := &ssh.ClientConfig{
		User:            user,
		Timeout:         30 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
	}

	conn, err := ssh.Dial("tcp", host+":"+string(port), config)
	if err != nil {
		return nil, err
	}

	session, err := conn.NewSession()
	if err != nil {
		return nil, err
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err = session.RequestPty("xterm", 80, 40, modes); err != nil {
		return nil, err
	}

	return session, nil
}

//Do do commandlist
func Do(user, password, host string, cmds []string, port, timeout int, ch chan Result) {
	res := Result{
		Host:    host,
		Success: false,
	}
	channel := make(chan Result)

	session, err := session(user, host, password, port)
	if err != nil {
		res.Result = err.Error()
		ch <- res
		return
	}
	defer session.Close()

	cmds = append(cmds, "exit")

	stdin, _ := session.StdinPipe()
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if err = session.Shell(); err != nil {
		res.Result = err.Error()
		ch <- res
		return
	}

	go func() {
		for _, cmd := range cmds {
			stdin.Write([]byte(cmd + "\n"))
		}

		session.Wait()

		if stderr.String() != "" {
			res.Result = stderr.String()
		} else {
			res.Success = true
			res.Result = stdout.String()
		}

		channel <- res
	}()

	select {
	case <-time.After(time.Duration(timeout) * time.Second):
		res.Result = ("SSH run timeout：" + strconv.Itoa(timeout) + " second.")
		ch <- res
	case res = <-channel:
		ch <- res
	}

	return
}
