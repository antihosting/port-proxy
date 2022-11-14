/**
  Copyright (c) 2022 Zander Schwid & Co. LLC. All rights reserved.
*/

package main

import (
	"context"
	proxy "github.com/antihosting/port-proxy"
	"github.com/pkg/errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	rt "runtime"
	"time"
)

var (
	Exec    string
	Version string
	Build   string
	HashedToken   string
)

func main() {

	rt.GOMAXPROCS(rt.NumCPU())

	Exec = os.Args[0]

	os.Exit(Run(os.Args[1:]))

}

func Run(args []string) int {

	if err := doRun(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	} else {
		return 0
	}

}

func doRun(args []string) error {

	if len(args) == 0 {
		flag.PrintDefaults()
		return errors.New("empty flags")
	}

	flag.CommandLine.Parse(args)

	if *GenerateToken {
		token, err := proxy.GenerateToken()
		if err != nil {
			return err
		}
		hashedToken, err := proxy.HashToken(token)
		if err != nil {
			return err
		}
		fmt.Printf("Token: %s\n", token)
		fmt.Printf("Hashed Token: %s\n", hashedToken)
		return nil
	}

	token := os.Getenv("PORT_PROXY_TOKEN")
	if token == "" {
		token = proxy.PromptPassword("Enter token: ")
	}

	hashedToken, err := proxy.HashToken(token)
	if err != nil {
		return err
	}

	if hashedToken != HashedToken {
		return errors.New("invalid token")
	}

	if len(Ports) == 0 {
		return errors.New("empty forward ports")
	}

	readTimeout, err := time.ParseDuration(*ReadTimeout)
	if err != nil {
		return errors.Errorf("incorrect read timeout '%s', %v", *ReadTimeout, err)
	}

	writeTimeout, err := time.ParseDuration(*WriteTimeout)
	if err != nil {
		return errors.Errorf("incorrect write timeout '%s', %v", *WriteTimeout, err)
	}

	if *BenchmarkTest == "http" {
		return proxy.RunHttpBenchmarkTest(*ListenIP, Ports[0], *BenchmarkSize, *Count)
	}

	if *BenchmarkTest == "socket" {
		return proxy.RunSocketBenchmarkTest(*ListenIP, Ports[0], *BenchmarkSize, *Count)
	}

	if !*Foreground {
		// fork the process to run in background
		return startBackground(token)
	}

	var logFile *os.File
	var logWriter io.Writer

	if *LogFile == "stdout" {
		logWriter = os.Stdout
	} else {
		var err error
		logFile, err = os.OpenFile(*LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return errors.Errorf("fail to open file '%s', %v", *LogFile, err)
		}
		logWriter = logFile
	}
	defer func() {
		if logFile != nil {
			logFile.Close()
		}
	}()

	log := log.New(logWriter,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	log.Printf("%s %s %s\n", Exec, Version, Build)
	log.Printf("Listen IP Address: %s\n", *ListenIP)
	log.Printf("Forward Ports: %+v\n", Ports)
	log.Printf("Verbose: %v\n", *Verbose)

	ctx := context.WithValue(context.Background(), proxy.ReadTimeoutKey{}, readTimeout)
	ctx = context.WithValue(context.Background(), proxy.WriteTimeoutKey{}, writeTimeout)

	return proxy.RunProxy(ctx, *ListenIP, Ports, log, *Verbose)
}

