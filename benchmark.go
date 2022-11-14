/**
  Copyright (c) 2022 Zander Schwid & Co. LLC. All rights reserved.
*/

package main

import (
	"bytes"
	"context"
	"github.com/pkg/errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

func RunSocketBenchmarkTest(ip string, forward ForwardPort, bs, count int) error {

	listenAddr := fmt.Sprintf("%s:%d", ip, forward.SrcPort)
	forwardAddr := fmt.Sprintf("%s:%d", ip, forward.DstPort)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	echo := NewEchoServer(ctx, forwardAddr)
	err := echo.Bind()
	if err != nil {
		return err
	}
	defer echo.Close()

	go echo.Serve()

	go RunProxy(ctx, ip, []ForwardPort { forward }, log.Default(), false)

	return runSocketBenchmark(ctx, listenAddr, bs, count)

}

func runSocketBenchmark(ctx context.Context, listenAddr string, bs, count int) error {

	payload := make([]byte, bs)
	answer := make([]byte, len(payload))

	var sum float64
	start := time.Now()

	conn, err := net.Dial("tcp", listenAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	for i := 0; i < count; i++ {

		rstart := time.Now()

		g, _ := errgroup.WithContext(ctx)
		g.Go(func() error {
			_, err = io.Copy(conn, bytes.NewReader(payload))
			return err
		})

		g.Go(func() error {
			_, err := io.ReadFull(conn, answer)
			if !bytes.Equal(payload, answer) {
				return errors.New("payload not equal to answer")
			}
			return err
		})

		err = g.Wait()
		if err != nil {
			return errors.Errorf("errgroup err, %v", err)
		}

		relapsed := time.Now().Sub(rstart)
		relapsedMillis := float64(relapsed) / float64(time.Millisecond)
		sum += relapsedMillis

	}
	elapsed := time.Now().Sub(start)
	elapsedSec := float64(elapsed) / float64(time.Second)

	totalBytes := int64(bs) * int64(count)

	throughputBytes := float64(totalBytes) / elapsedSec
	throughputMB := throughputBytes / float64(1 << 20)

	latencyMillis := sum / float64(count)

	log.Printf("Latency %0.4f Millis\n", latencyMillis)
	log.Printf("Throughput %0.2f MB\n", throughputMB)
	return nil

}

func RunHttpBenchmarkTest(ip string, forward ForwardPort, bs, count int) error {

	listenAddr := fmt.Sprintf("%s:%d", ip, forward.SrcPort)
	forwardAddr := fmt.Sprintf("%s:%d", ip, forward.DstPort)

	server := &http.Server{
		Addr:    forwardAddr,
		Handler: &echoHandler{},
	}
	defer server.Close()

	go server.ListenAndServe()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go RunProxy(ctx, ip, []ForwardPort { forward }, log.Default(), false)

	return runHttpBenchmark(listenAddr, bs, count)
}

func runHttpBenchmark(listenAddr string, bs, count int) error {

	payload := make([]byte, bs)

	var sum float64

	url := fmt.Sprintf("http://%s", listenAddr)
	start := time.Now()
	for i := 0; i < count; i++ {

		rstart := time.Now()

		resp, err := http.Post(url, "", bytes.NewReader(payload))
		if err != nil {
			return err
		}

		answer, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		relapsed := time.Now().Sub(rstart)
		relapsedMillis := float64(relapsed) / float64(time.Millisecond)
		sum += relapsedMillis

		if !bytes.Equal(payload, answer) {
			return errors.New("payload not equal to answer")
		}

	}
	elapsed := time.Now().Sub(start)
	elapsedSec := float64(elapsed) / float64(time.Second)

	totalBytes := int64(bs) * int64(count)

	throughputBytes := float64(totalBytes) / elapsedSec
	throughputMB := throughputBytes / float64(1 << 20)

	latencyMillis := sum / float64(count)

	log.Printf("Latency %0.4f Millis\n", latencyMillis)
	log.Printf("Throughput %0.2f MB\n", throughputMB)
	return nil

}
