package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/robotomize/powwy/internal/logging"
	"github.com/robotomize/powwy/pkg/proto"
)

type (
	HandleFunc     func(r proto.Request, w *proto.ResponseWriter)
	MiddlewareFunc func(next HandleFunc) HandleFunc
)

var (
	ErrUnknownCommand = errors.New("unknown command")
	ErrInternalServer = errors.New("internal server error")
)

type ContextSubjectKey string

const ClientSubjectCtxKey = ContextSubjectKey("client_addr")

func New(l net.Listener, connDeadline time.Duration) (*Server, error) {
	return &Server{
		listener:     l,
		quit:         make(chan struct{}, 1),
		handlers:     make(map[string]HandleFunc),
		connDeadline: connDeadline,
	}, nil
}

type Server struct {
	connDeadline time.Duration
	listener     net.Listener
	wg           sync.WaitGroup
	quit         chan struct{}

	mtx      sync.RWMutex
	handlers map[string]HandleFunc
}

func (s *Server) HandleFunc(cmd string, f HandleFunc) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.handlers[cmd] = f
}

func (s *Server) Serve(ctx context.Context) error {
	logger := logging.FromContext(ctx).Named("Serve")

	s.wg.Add(1)

	errCh := make(chan error, 1)
	go func() {
		<-ctx.Done()

		logger.Debug("Server.Serve: context closed")
		logger.Debug("Server.Serve: shutting down")

		s.wg.Done()
		if err := s.close(); err != nil {
			logger.Debugf("listener close: %v", err)
		}
	}()

	logger.Debug("Server.ServerTCP: started")
	clntCtx, cancelConn := context.WithCancel(ctx)

AcceptLoop:
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.quit:
				cancelConn()
				break AcceptLoop
			default:
			}

			logger.Errorf("accept conn: %v", err)
			continue
		}

		s.wg.Add(1)

		go s.read(clntCtx, conn)
	}

	s.wg.Wait()
	logger.Debug("Server.Serve: serving stopped")

	select {
	case err := <-errCh:
		return fmt.Errorf("failed to shutdown: %w", err)
	default:
		return nil
	}
}

func (s *Server) read(ctx context.Context, conn net.Conn) {
	defer s.wg.Done()
	logger := logging.FromContext(ctx).Named("ReadConn")

	logger.Debugf("client %s connected", conn.RemoteAddr().String())

	rw := proto.NewResponseWriter(conn)

	ctx, cancel := context.WithCancel(ctx)

	go func() {
		if err := rw.ReadAll(ctx); err != nil {
			logger.Errorf("ReadAll: %v", err)
		}
	}()

	defer func() {
		if err := rw.Close(); err != nil {
			logger.Errorf("close conn: %v", err)
		}

		logger.Debugf("powwy %s disconnected", conn.RemoteAddr().String())
	}()

	go func() {
		<-ctx.Done()
		if !rw.IsClosed() {
			if err := conn.SetDeadline(time.Now().Add(s.connDeadline)); err != nil {
				logger.Errorf("conn set deadline: %v", err)
				return
			}
		}
	}()

	defer cancel()

	for request := range rw.ReadMessage() {
		c := context.WithValue(
			request.Context(), ClientSubjectCtxKey, strings.Replace(
				conn.RemoteAddr().String(), ":", "@", -1),
		)
		request = request.WithContext(c)

		if proto.IsAvailableCommand(request.Cmd) {
			switch request.Cmd {
			case proto.DISC:
				if _, err := rw.SendOK(); err != nil {
					logger.Errorf("send err: %v", err)
				}

				return
			}

			if err := s.handle(ctx, request, rw); err != nil {
				logger.Errorf("handle: %v", err)
			}

			continue
		}

		if _, err := rw.SendErr(ErrUnknownCommand.Error()); err != nil {
			logger.Errorf("send err: %v", err)
		}
	}
}

func (s *Server) handle(ctx context.Context, request proto.Request, w *proto.ResponseWriter) error {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	logger := logging.FromContext(ctx)

	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("panic: %v", err)
		}
	}()

	f, ok := s.handlers[request.Cmd]
	if !ok {
		if _, err := w.SendErr(ErrInternalServer.Error()); err != nil {
			return fmt.Errorf("send err: %w", err)
		}

		return nil
	}

	f(request, w)

	return nil
}

func (s *Server) close() error {
	close(s.quit)

	if err := s.listener.Close(); err != nil {
		return fmt.Errorf("listener close: %w", err)
	}

	return nil
}
