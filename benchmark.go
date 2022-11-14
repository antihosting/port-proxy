/**
  Copyright (c) 2022 Zander Schwid & Co. LLC. All rights reserved.
*/

package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func RunBenchmarkTest(ip string, forward ForwardPort, bs, count int) error {

	listenAddr := fmt.Sprintf("%s:%d", ip, forward.SrcPort)
	forwardAddr := fmt.Sprintf("%s:%d", ip, forward.DstPort)

	server := &http.Server{
		Addr:    forwardAddr,
		Handler: &echoHandler{},
	}
	defer server.Close()

	go server.ListenAndServe()

	payload := make([]byte, bs)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go RunProxy(ctx, ip, []ForwardPort { forward }, log.Default(), false)

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
