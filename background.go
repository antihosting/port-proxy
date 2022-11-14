/**
  Copyright (c) 2022 Zander Schwid & Co. LLC. All rights reserved.
*/

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

func startBackground(token string) error {

	executable, err := os.Executable()
	if err != nil {
		return err
	}

	args := []string{
		"-f",
		"-ip", *ListenIP,
		"-log", executable + ".log",
		"-srt", *ReadTimeout,
		"-swt", *WriteTimeout,
	}

	for _, portForward := range Ports {
		args = append(args, "-p", fmt.Sprintf("%d:%d", portForward.SrcPort, portForward.DstPort))
	}

	if *Verbose {
		args = append(args, "-v")
	}

	cmd := exec.Command(executable, args...)
	fmt.Printf("Run cmd: %v\n", cmd)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	defer stdin.Close()

	io.WriteString(stdin, token+"\n")

	if err := cmd.Start(); err != nil {
		return err
	}

	fmt.Println("Daemon process ID is : ", cmd.Process.Pid)

	content := fmt.Sprintf("%d", cmd.Process.Pid)
	ioutil.WriteFile(executable+".pid", []byte(content), 0660)

	fmt.Println("Proxy started in background.")
	return nil
}
