package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
)

// Response ...
type Response struct {
	Output bytes.Buffer
	Logs   bytes.Buffer
}

func (l *Lifecycle) execute(command string) (Response, error) {
	cmd := exec.Command("bash", "-c", fmt.Sprintf(command))
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if errb.Len() > 0 {
		// All logs are currently directed to stderr
		fmt.Println(errb.String())
	}
	return Response{Output: outb, Logs: errb}, err
}

func (res Response) findInLogs(regex string) (string, error) {
	r, err := regexp.Compile(regex)
	if err != nil {
		return "", err
	}

	return r.FindString(res.Logs.String()), nil
}
