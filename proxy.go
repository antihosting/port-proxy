/**
  Copyright (c) 2022 Zander Schwid & Co. LLC. All rights reserved.
*/

package proxy

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type ForwardPort struct {
	SrcPort int
	DstPort int
}

func (t ForwardPort) String() string {
	return fmt.Sprintf("%d:%d", t.SrcPort, t.DstPort)
}

func RunProxy(ctx context.Context, ip string, ports []ForwardPort, log *log.Logger, verbose bool) (err error) {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var serverList []*proxyServer

	for _, forward := range ports {

		listenAddr := fmt.Sprintf("%s:%d", ip, forward.SrcPort)
		forwardAddr := fmt.Sprintf("%s:%d", ip, forward.DstPort)
		server := NewProxyServer(ctx, listenAddr, forwardAddr, log, verbose)

		serverList = append(serverList, server)
	}

	var bindErrors []error
	for _, server := range serverList {

		if err := server.Bind(); err != nil {
			log.Printf("Bind server %v error, %v\n", server, err)
			bindErrors = append(bindErrors, err)
		}

	}

	if len(bindErrors) > 0 {
		closeAll(serverList, log)
		return errors.Errorf("bind errors: %+v", bindErrors)
	}

	cnt := 0
	g, _ := errgroup.WithContext(ctx)

	for _, server := range serverList {
		g.Go(server.Serve)
		cnt++
	}
	log.Printf("Daemon started with %d proxy servers\n", cnt)

	go func() {

		signalCh := make(chan os.Signal, 1)
		signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

		var signal os.Signal

		select {
		case signal = <- signalCh:
		case <- ctx.Done():
			signal = syscall.SIGABRT
		}

		log.Printf("Daemon stopped by signal %s\n", signal.String())
		cancel()
		closeAll(serverList, log)
	}()

	return g.Wait()
}

func closeAll(serverList []*proxyServer, log *log.Logger) error {

	for _, server := range serverList {
		if err := server.Close(); err != nil {
			log.Printf("Close server %v error, %v\n", server, err)
		}
	}

	return nil
}