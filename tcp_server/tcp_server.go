package tcp_server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrServerClosed     = errors.New("tcp: Service closed")
	ErrAbortHandler     = errors.New("tcp: abort TCPHandler")
	ServiceContextKey   = &contextKey{"tcp-server"}
	LocalAddrContextKey = &contextKey{"local-addr"}
)

type onceCloseListener struct {
	net.Listener
	once     sync.Once
	closeErr error
}

func (oc *onceCloseListener) Close() error {
	oc.once.Do(oc.close)
	return oc.closeErr
}

func (oc *onceCloseListener) close() {
	oc.closeErr = oc.Listener.Close()
}

type TCPHandler interface {
	ServeTCP(ctx context.Context, conn net.Conn)
}

type TcpServer struct {
	Addr    string     //监听地址
	Handler TCPHandler //实际逻辑回调设置handler
	err     error
	BaseCtx context.Context //上下文
	//读写超时设置
	WriteTimeout     time.Duration
	ReadTimeout      time.Duration
	KeepAliveTimeout time.Duration //连接一直保持 发数据包的时间
	mu               sync.Mutex    //锁
	inShutdown       int32         //是否关闭
	doneChan         chan struct{}
	l                *onceCloseListener //单次启动时，listen需要执行的一次性操作的设置
}

func (srv *TcpServer) shuttingDown() bool {
	//原子方法验证是否为0
	return atomic.LoadInt32(&srv.inShutdown) != 0
}

func (srv *TcpServer) ListenAndServe() error {
	//验证服务是否关闭
	if srv.shuttingDown() {
		return ErrServerClosed
	}
	if srv.doneChan == nil {
		srv.doneChan = make(chan struct{})
	}
	addr := srv.Addr
	if addr == "" {
		return errors.New("need addr")
	}
	//调用核心的官方方法，并传入到Serve中
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return srv.Serve(tcpKeepAliveListener{ln.(*net.TCPListener)})
}

func (srv *TcpServer) Close() error {
	atomic.StoreInt32(&srv.inShutdown, 1)
	close(srv.doneChan) //关闭channel
	srv.l.Close()       //执行listener关闭
	return nil
}

func (srv *TcpServer) getDoneChan() <-chan struct{} {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if srv.doneChan == nil {
		srv.doneChan = make(chan struct{})
	}
	return srv.doneChan
}

func (srv *TcpServer) newConn(rwc net.Conn) *conn {
	c := &conn{
		server: srv,
		rwc:    rwc,
	}
	// 设置超时时间参数返回
	if d := c.server.ReadTimeout; d != 0 {
		c.rwc.SetReadDeadline(time.Now().Add(d))
	}
	if d := c.server.WriteTimeout; d != 0 {
		c.rwc.SetWriteDeadline(time.Now().Add(d))
	}
	if d := c.server.KeepAliveTimeout; d != 0 {
		if tcpConn, ok := c.rwc.(*net.TCPConn); ok {
			tcpConn.SetKeepAlive(true)
			tcpConn.SetKeepAlivePeriod(d)
		}
	}
	return c
}

func (srv *TcpServer) Serve(l net.Listener) error {
	srv.l = &onceCloseListener{Listener: l}
	//退出时执行listener关闭
	defer srv.l.Close()
	if srv.BaseCtx == nil {
		srv.BaseCtx = context.Background()
	}
	BaseCtx := srv.BaseCtx
	ctx := context.WithValue(BaseCtx, ServiceContextKey, srv)
	//轮询Listener的Accept方法，获取客户端发来的conn
	for {
		rw, err := l.Accept()
		if err != nil {
			select {
			case <-srv.getDoneChan():
				return ErrServerClosed
			default:
			}
			fmt.Printf("accept fail, err: %v\n", err)
			continue
		}
		//拿到便马上创建连接
		c := srv.newConn(rw)
		go c.serve(ctx)
	}
	return nil
}
