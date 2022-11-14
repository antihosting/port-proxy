package main

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"
)

func TestProxy(t *testing.T) {

	server := &http.Server{
		Addr:              "127.0.0.1:50551",
		Handler: &echoHandler{},
	}

	go server.ListenAndServe()
	
	var ports []ForwardPort

	ports = append(ports, ForwardPort {
		SrcPort: 50550,
		DstPort: 50551,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go RunProxy(ctx, "127.0.0.1", ports, log.Default(), false)

	payload := make([]byte, 1 << 20)

	start := time.Now()

	resp, err := http.Post("http://127.0.0.1:50550", "", bytes.NewReader(payload))
	require.NoError(t, err)

	answer, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	elapsed := time.Now().Sub(start)

	require.True(t, bytes.Equal(payload, answer))

	println(elapsed / time.Millisecond)

}
