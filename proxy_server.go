/**
  Copyright (c) 2022 Zander Schwid & Co. LLC. All rights reserved.
*/

package proxy

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"log"
	"net"
	"go.uber.org/atomic"
	"strings"
	"sync"
	"time"
)

// value time.Duration
type ReadTimeoutKey struct {
}

// value time.Duration
type WriteTimeoutKey struct {
}

type proxyServer struct {

	ctx context.Context

	listenAddr string
	lc         net.ListenConfig
	listener   net.Listener

	cancelFn    context.CancelFunc

	forwardAddr string

	log      *log.Logger
	verbose  bool

	readTimeout  time.Duration
	writeTimeout time.Duration

	running    atomic.Bool
	closeOnce  sync.Once
}

func NewProxyServer(ctx context.Context, listenAddr, forwardAddr string, log *log.Logger, verbose bool) *proxyServer {
	return &proxyServer{
		ctx: ctx,
		listenAddr: listenAddr,
		forwardAddr: forwardAddr,
		log: log,
		verbose: verbose,
	}
}

func (t *proxyServer) String() string {
	return fmt.Sprintf("ProxyServer {%s to %s}", t.listenAddr, t.forwardAddr)
}

func (t *proxyServer) Bind() (err error) {

	t.listener, err = t.lc.Listen(t.ctx, "tcp4", t.listenAddr)
	if err != nil {
		return errors.Errorf("Listen address is busy '%s', %v", t.listenAddr, err)
	}

	return nil
}

func (t *proxyServer) Serve() (err error) {

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

	t.log.Printf("ProxyServe Started '%s' -> '%s'\n", t.listenAddr, t.forwardAddr)

	t.running.Store(true)
	err = t.doServe(serveCtx)
	t.running.Store(false)

	t.cancelFn()

	if err != nil && strings.Contains(err.Error(), "closed") {
		err = nil
	}

	t.log.Printf("ProxyServe Ended '%s' -> '%s' with error %v\n", t.listenAddr, t.forwardAddr, err)
	return err
}

func (t *proxyServer) doServe(ctx context.Context) error {
	for t.running.Load() {
		conn, err := t.listener.Accept()
		if err != nil {
			return err
		}
		go t.serveConn(ctx, conn)
	}
	return nil
}

func (t *proxyServer) serveConn(ctx context.Context, conn net.Conn) error {
	defer conn.Close()

	if d, ok := ctx.Value(ReadTimeoutKey{}).(time.Duration); ok && d != 0 {
		conn.SetReadDeadline(time.Now().Add(d))
	}

	if d, ok := ctx.Value(WriteTimeoutKey{}).(time.Duration); ok && d != 0 {
		conn.SetWriteDeadline(time.Now().Add(d))
	}

	return t.forward(ctx, conn, t.forwardAddr)
}

func (t *proxyServer) Close() (err error) {
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

func (t *proxyServer) forward(ctx context.Context, conn net.Conn, destAddr string) error {

	target, err := net.Dial("tcp", destAddr)
	if err != nil {
		return err
	}
	defer target.Close()

	// Start proxying
	s2cCh := proxy(target, conn)
	c2sCh := proxy(conn, target)
	var total int64

	defer func() {

		if t.verbose {
			t.log.Printf("Traffic from '%s' to '%s' amount %d\n", conn.RemoteAddr().String(),  target.RemoteAddr().String(), total)
		}

		go func() {
			// make sure that all goroutings already finished
			time.Sleep(time.Millisecond * 100)
			close(s2cCh)
			close(c2sCh)
		}()
	}()

	// We don't know which side is going to stop sending first, so we need a select between the two.
	for i := 0; i < 2; i++ {
		select {
		case <- ctx.Done():
			target.Close()
			conn.Close()
			i++
		case s2c := <-s2cCh:
			total += s2c.Cnt
			if s2c.Err == io.EOF {
				break // select, continue with client
			}
			if s2c.Err != nil {
				return errors.Errorf("server closed connection with error: %v", s2c.Err)
			}
		case c2s := <-c2sCh:
			total += c2s.Cnt
			if c2s.Err == io.EOF {
				break // select, continue with server
			}
			if c2s.Err != nil {
				return errors.Errorf("client closed connection with error: %v", c2s.Err)
			}
		}
	}
	return nil

}

type proxyResult struct {
	Cnt int64
	Err error
}

type closeWriter interface {
	CloseWrite() error
}

// proxy is used to suffle data from src to destination, and sends errors
// down to dedicated channel
func proxy(dst io.Writer, src io.Reader) chan proxyResult {
	ret := make(chan proxyResult, 1)
	go func() {
		cnt, err := io.Copy(dst, src)
		//if tcpConn, ok := dst.(closeWriter); ok {
		//	tcpConn.CloseWrite()
		//}
		ret <- proxyResult{cnt, err}
	}()
	return ret
}

