package tcp_service

import (
	"context"
	"fmt"
	"net"
	"runtime"
)

type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (net.Conn, error) {
	tcp, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}
	return tcp, nil
}

type contextKey struct {
	name string
}

func (k *contextKey) String() string {
	return "tcp_proxy context value " + k.name
}

type conn struct {
	server     *TCPServer
	cancelCtx  context.CancelFunc
	rwc        net.Conn
	remoteAddr string
}

func (c *conn) close() {
	c.rwc.Close()
}

func (c conn) serve(ctx context.Context) {
	defer func() {
		if err := recover(); err != nil && err != ErrAbortHandler {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			fmt.Printf("tcp: panic serving %v: %v\n%s", c.remoteAddr, err, buf)
		}
		c.close()
	}()
	c.remoteAddr = c.rwc.RemoteAddr().String()
	ctx = context.WithValue(ctx, LocalAddrContextKey, c.rwc.LocalAddr())
	if c.server.Handler == nil {
		panic("handler empty")
	}
	c.server.Handler.ServeTCP(ctx,c.rwc)
}
