/**
  Copyright (c) 2022 Zander Schwid & Co. LLC. All rights reserved.
*/


package proxy_test

import (
	"bytes"
	"context"
	"errors"
	proxy "github.com/antihosting/port-proxy"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestHTTPProxy(t *testing.T) {

	bs := 1 << 20
	count := 1024

	server := &http.Server{
		Addr:              "127.0.0.1:50551",
		Handler: &proxy.EchoHandler{},
	}

	go server.ListenAndServe()
	
	var ports []proxy.ForwardPort

	ports = append(ports, proxy.ForwardPort{
		SrcPort: 50550,
		DstPort: 50551,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go proxy.RunProxy(ctx, "127.0.0.1", ports, log.Default(), false)

	payload := make([]byte, bs)

	time.Sleep(time.Millisecond)

	var sum float64
	start := time.Now()

	for i := 0; i < count; i++ {

		rstart := time.Now()

		resp, err := http.Post("http://127.0.0.1:50550", "", bytes.NewReader(payload))
		require.NoError(t, err)

		answer, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.True(t, bytes.Equal(payload, answer))

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

}

func TestSocketProxy(t *testing.T) {

	bs := 1 << 20
	count := 1024

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	echo := proxy.NewEchoServer(ctx, "127.0.0.1:50451")
	err := echo.Bind()
	require.NoError(t, err)

	defer echo.Close()
	go echo.Serve()

	var ports []proxy.ForwardPort

	ports = append(ports, proxy.ForwardPort{
		SrcPort: 50450,
		DstPort: 50451,
	})

	go proxy.RunProxy(ctx, "127.0.0.1", ports, log.Default(), false)

	time.Sleep(time.Millisecond)

	payload := make([]byte, bs)
	answer := make([]byte, len(payload))

	var sum float64
	start := time.Now()

	conn, err := net.Dial("tcp", "127.0.0.1:50450")
	require.NoError(t, err)
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
		require.NoError(t, err)

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

}

