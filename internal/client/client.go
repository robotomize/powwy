package client

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/robotomize/powwy/pkg/hashcash"
	"github.com/robotomize/powwy/pkg/proto"
)

var (
	ErrWrongAnswer = errors.New("wrong answer")
	ErrWongPayload = errors.New("payload wrong")
)

type ProtError struct {
	OriginMessage proto.Request
	Message       string
}

func (e ProtError) Error() string {
	return e.Message
}

type Config struct {
	Addr    string
	Network string
}

func NewClient(config Config) *Client {
	return &Client{config: config, inbound: make(chan proto.Request, 1)}
}

type Client struct {
	config  Config
	inbound chan proto.Request
	rw      *proto.ResponseWriter
	sync.Once
}

func (c *Client) Connect(ctx context.Context) error {
	return c.connect(ctx)
}

func (c *Client) connect(ctx context.Context) error {
	conn, err := net.Dial(c.config.Network, c.config.Addr)
	if err != nil {
		return fmt.Errorf("net.Dial: %w", err)
	}

	c.rw = proto.NewResponseWriter(conn)

	go c.read(ctx)

	return nil
}

func (c *Client) read(ctx context.Context) {
	defer close(c.inbound)

	go func() {
		if err := c.rw.ReadAll(ctx); err != nil {
			fmt.Printf("ReadAll: %v", err)
		}
	}()

	for request := range c.rw.ReadMessage() {
		if !proto.IsAvailableCommand(request.Cmd) {
			continue
		}

		c.inbound <- request
	}
}

func (c *Client) SendREQ(ctx context.Context) (hashcash.Header, error) {
	if err := c.connect(ctx); err != nil {
		return hashcash.Header{}, err
	}

	if _, err := c.rw.SendREQ(); err != nil {
		return hashcash.Header{}, fmt.Errorf("send req: %w", err)
	}

	select {
	case <-ctx.Done():
		return hashcash.Header{}, fmt.Errorf("ctx done: %w", ctx.Err())
	case request, ok := <-c.inbound:
		if !ok {
			return hashcash.Header{}, fmt.Errorf("server unexpectedly closed connection")
		}

		if request.Cmd == proto.ERR {
			return hashcash.Header{}, ProtError{
				OriginMessage: request,
				Message:       string(request.Body),
			}
		}

		if request.Cmd != proto.RSV {
			return hashcash.Header{}, ProtError{
				OriginMessage: request,
				Message:       ErrWrongAnswer.Error(),
			}
		}

		header, err := hashcash.Parse(string(request.Body))
		if err != nil {
			return hashcash.Header{}, ProtError{
				OriginMessage: request,
				Message:       ErrWongPayload.Error(),
			}
		}

		return header, nil
	}
}

func (c *Client) SendRES(ctx context.Context, msg string) ([]byte, error) {
	if err := c.connect(ctx); err != nil {
		return nil, err
	}

	if _, err := c.rw.SendRES(msg); err != nil {
		return nil, fmt.Errorf("send req: %w", err)
	}

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("ctx done: %w", ctx.Err())
	case request, ok := <-c.inbound:
		if !ok {
			return nil, fmt.Errorf("server unexpectedly closed connection")
		}

		if request.Cmd == proto.ERR {
			return nil, ProtError{
				OriginMessage: request,
				Message:       string(request.Body),
			}
		}

		if request.Cmd != proto.RSV {
			return nil, ProtError{
				OriginMessage: request,
				Message:       ErrWrongAnswer.Error(),
			}
		}

		return request.Body, nil
	}
}

func (c *Client) SendDISC(ctx context.Context) error {
	if err := c.connect(ctx); err != nil {
		return err
	}
	if _, err := c.rw.SendDISC(); err != nil {
		return fmt.Errorf("send req: %w", err)
	}

	select {
	case <-ctx.Done():
		return fmt.Errorf("ctx done: %w", ctx.Err())
	case request, ok := <-c.inbound:
		if !ok {
			return fmt.Errorf("server unexpectedly closed connection")
		}

		if request.Cmd == proto.ERR {
			return ProtError{
				OriginMessage: request,
				Message:       string(request.Body),
			}
		}

		if request.Cmd != proto.OK {
			return ProtError{
				OriginMessage: request,
				Message:       ErrWrongAnswer.Error(),
			}
		}
	}

	return nil
}
