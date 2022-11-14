/**
  Copyright (c) 2022 Zander Schwid & Co. LLC. All rights reserved.
*/

package proxy

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"go.uber.org/atomic"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type EchoHandler struct {
}

func(t *EchoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

type echoServer struct {

	ctx context.Context

	listenAddr string
	lc         net.ListenConfig
	listener   net.Listener

	cancelFn    context.CancelFunc

	running    atomic.Bool
	closeOnce  sync.Once
}

func NewEchoServer(ctx context.Context, listenAddr string) *echoServer {
	return &echoServer{
		ctx: ctx,
		listenAddr: listenAddr,
	}
}


func (t *echoServer) String() string {
	return fmt.Sprintf("EchoServer {%s}", t.listenAddr)
}

func (t *echoServer) Bind() (err error) {

	t.listener, err = t.lc.Listen(t.ctx, "tcp4", t.listenAddr)
	if err != nil {
		return errors.Errorf("Listen address is busy '%s', %v", t.listenAddr, err)
	}

	return nil
}

func (t *echoServer) Serve() (err error) {

	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				err = v
			case string:
				err = errors.New(v)
			default:
				err = errors.Errorf("%v", v)
			}
		}
	}()

	var serveCtx context.Context
	serveCtx, t.cancelFn = context.WithCancel(t.ctx)

	t.running.Store(true)
	err = t.doServe(serveCtx)
	t.running.Store(false)

	t.cancelFn()

	if err != nil && strings.Contains(err.Error(), "closed") {
		err = nil
	}

	return err
}

func (t *echoServer) doServe(ctx context.Context) error {
	for t.running.Load() {
		conn, err := t.listener.Accept()
		if err != nil {
			return err
		}
		go t.serveConn(ctx, conn)
	}
	return nil
}

func (t *echoServer) serveConn(ctx context.Context, conn net.Conn) error {
	defer conn.Close()

	if d, ok := ctx.Value(ReadTimeoutKey{}).(time.Duration); ok && d != 0 {
		conn.SetReadDeadline(time.Now().Add(d))
	}

	if d, ok := ctx.Value(WriteTimeoutKey{}).(time.Duration); ok && d != 0 {
		conn.SetWriteDeadline(time.Now().Add(d))
	}

	_, err := io.Copy(conn, conn)
	return err
}

func (t *echoServer) Close() (err error) {
	t.running.Store(false)

	t.closeOnce.Do(func() {

		if t.listener != nil {
			err = t.listener.Close()
			if err != nil && strings.Contains(err.Error(), "closed") {
				err = nil
			}
		}

	})

	return nil
}

