package proto

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
)

// Proto specification tags
const (
	REQ  = "REQ"  // REQ - request challenge
	RES  = "RES"  // RES - request resource
	RSV  = "RSV"  // RSV - response with payload
	OK   = "OK"   // OK - command accepted
	ERR  = "ERR"  // ERR - command err
	DISC = "DISC" // DISC - initialize close connection
)

const (
	PayloadDelimiter = "|"
	TCPDelimiter     = '\r'
)

var ErrBodyLen = errors.New("body length incorrect")

var _ io.Writer = (*ResponseWriter)(nil)

func IsAvailableCommand(c string) bool {
	for _, cmd := range []string{REQ, RES, RSV, OK, ERR, DISC} {
		if c == cmd {
			return true
		}
	}

	return false
}

func NewResponseWriter(rwc io.ReadWriteCloser) *ResponseWriter {
	return &ResponseWriter{
		ReadWriteCloser: rwc,
		inbound:         make(chan Request, 1),
	}
}

func NewRequest(ctx context.Context, cmd string, body []byte) Request {
	return Request{Cmd: cmd, Body: body, ctx: ctx}
}

type Request struct {
	Cmd  string
	Body []byte
	ctx  context.Context
}

func (r Request) WithContext(ctx context.Context) Request {
	r2 := Request{
		Cmd:  r.Cmd,
		Body: make([]byte, len(r.Body)),
		ctx:  ctx,
	}

	copy(r2.Body, r.Body)

	return r2
}

func (r Request) Context() context.Context {
	return r.ctx
}

type ResponseWriter struct {
	io.ReadWriteCloser

	inbound chan Request

	mtx    sync.RWMutex
	closed bool
}

func (t *ResponseWriter) ReadMessage() <-chan Request {
	return t.inbound
}

func (t *ResponseWriter) Close() error {
	t.mtx.Lock()
	t.closed = true
	t.mtx.Unlock()

	return t.ReadWriteCloser.Close()
}

func (t *ResponseWriter) IsClosed() bool {
	t.mtx.RLock()
	defer t.mtx.RUnlock()

	return t.closed
}

func (t *ResponseWriter) ReadAll(ctx context.Context) error {
	defer close(t.inbound)

	for {
		b, err := bufio.NewReader(t.ReadWriteCloser).ReadBytes(TCPDelimiter)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}

			select {
			case <-ctx.Done():
				return nil
			default:
			}

			return fmt.Errorf("bufio.NewReader().ReadBytes: %w", err)
		}

		msg, err := t.parse(b)
		if err != nil {
			_, _ = t.SendErr(err.Error())
			continue
		}

		t.inbound <- msg
	}
}

func (t *ResponseWriter) Write(p []byte) (int, error) {
	b, cancel := GetBuf()
	defer cancel()

	b.Reset()
	b.Write(p)
	b.WriteByte(TCPDelimiter)

	return t.ReadWriteCloser.Write(b.Bytes())
}

func (t *ResponseWriter) SendRES(msg string) (int, error) {
	b, cancel := GetBuf()
	defer cancel()

	_, _ = fmt.Fprintf(b, "%s %d %s%s%c", RES, len(msg), PayloadDelimiter, msg, TCPDelimiter)

	return t.ReadWriteCloser.Write(b.Bytes())
}

func (t *ResponseWriter) SendREQ() (int, error) {
	b, cancel := GetBuf()
	defer cancel()

	b.WriteString(REQ)
	b.WriteByte(TCPDelimiter)

	return t.ReadWriteCloser.Write(b.Bytes())
}

func (t *ResponseWriter) SendErr(msg string) (int, error) {
	b, cancel := GetBuf()
	defer cancel()

	_, _ = fmt.Fprintf(b, "%s %d %s%s%c", ERR, len(msg), PayloadDelimiter, msg, TCPDelimiter)

	return t.ReadWriteCloser.Write(b.Bytes())
}

func (t *ResponseWriter) SendOK() (int, error) {
	b, cancel := GetBuf()
	defer cancel()

	b.WriteString(OK)
	b.WriteByte(TCPDelimiter)

	return t.ReadWriteCloser.Write(b.Bytes())
}

func (t *ResponseWriter) SendRSV(msg string) (int, error) {
	b, cancel := GetBuf()
	defer cancel()

	_, _ = fmt.Fprintf(b, "%s %d %s%s%c", RSV, len(msg), PayloadDelimiter, msg, TCPDelimiter)

	return t.ReadWriteCloser.Write(b.Bytes())
}

func (t *ResponseWriter) SendDISC() (int, error) {
	b, cancel := GetBuf()
	defer cancel()

	b.WriteString(DISC)
	b.WriteByte(TCPDelimiter)

	return t.ReadWriteCloser.Write(b.Bytes())
}

func (t *ResponseWriter) parse(p []byte) (Request, error) {
	cmd := bytes.ToUpper(bytes.TrimSpace(bytes.Split(p, []byte(" "))[0]))
	args := bytes.TrimSpace(bytes.TrimPrefix(p, cmd))
	var body []byte
	bodyLength := bytes.Split(args, []byte(PayloadDelimiter))[0]
	if string(cmd) == RES || string(cmd) == RSV || string(cmd) == ERR {
		length, err := strconv.Atoi(strings.Trim(string(bodyLength), " "))
		if err != nil {
			return Request{}, ErrBodyLen
		}

		if length == 0 {
			return Request{}, ErrBodyLen
		}

		padding := len(bodyLength) + len(PayloadDelimiter)
		if len(args) < padding+length {
			return Request{}, ErrBodyLen
		}

		body = args[padding : padding+length]
	}

	return NewRequest(context.Background(), string(cmd), body), nil
}
