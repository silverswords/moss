package src

import (
	"bytes"
	"fmt"
	"io/ioutil"
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

// session return a connection by user, host, password or key
// key is file path
// use password when key = ""
// when key not "", password is for key not for host
func session(user, host, password, key string, port int) (*ssh.Session, error) {
	config := &ssh.ClientConfig{
		User:            user,
		Timeout:         30 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		Config: ssh.Config{
			Ciphers: []string{"aes128-ctr", "aes192-ctr", "aes256-ctr", "aes128-gcm@openssh.com", "arcfour256", "arcfour128", "aes128-cbc", "3des-cbc", "aes192-cbc", "aes256-cbc"},
		},
	}

	//check key or password method
	if key != "" {
		pemBytes, err := ioutil.ReadFile(key)
		if err != nil {
			return nil, err
		}

		var signer ssh.Signer
		if password == "" {
			signer, err = ssh.ParsePrivateKey(pemBytes)
		} else {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(pemBytes, []byte(password))
		}
		if err != nil {
			return nil, err
		}

		config.Auth = append(config.Auth, ssh.PublicKeys(signer))
	}

	conn, err := ssh.Dial("tcp", host+":"+string(port), config)
	if err != nil {
		return nil, err
	}
	fmt.Println("a")
	session, err := conn.NewSession()
	if err != nil {
		return nil, err
	}
	fmt.Println("b")
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

// Do do commandlist by []string and add exit .
// use channel result struct src.Result include host,success,result string
func Do(user, password, host, key string, cmds []string, port, timeout int, ch chan Result) {
	res := Result{
		Host:    host,
		Success: false,
	}
	channel := make(chan Result)
	fmt.Println(2)
	session, err := session(user, host, password, key, port)
	if err != nil {
		res.Result = err.Error()
		ch <- res
		return
	}
	defer session.Close()

	cmds = append(cmds, "exit")
	fmt.Println(3)
	stdin, _ := session.StdinPipe()
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	//To-do: make function to error()
	if err = session.Shell(); err != nil {
		res.Result = err.Error()
		ch <- res
		return
	}

	go func() {
		for _, cmd := range cmds {
			stdin.Write([]byte(cmd + "\n"))
		}
		fmt.Println(4)
		session.Wait()

		if stderr.String() != "" {
			res.Result = stderr.String()
		} else {
			res.Success = true
			res.Result = stdout.String()
		}

		channel <- res
	}()
	fmt.Println(5)
	select {
	case <-time.After(time.Duration(timeout) * time.Second):
		res.Result = ("SSH run timeout：" + strconv.Itoa(timeout) + " second.")
		ch <- res
	case res = <-channel:
		ch <- res
	}

	return
}
