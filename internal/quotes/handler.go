package quotes

import (
	"fmt"

	"github.com/robotomize/powwy/internal/logging"
	"github.com/robotomize/powwy/internal/server"
	"github.com/robotomize/powwy/pkg/proto"
)

func NewHandler(quotes *Quotes, f HashFunc, g GenerateTokenFunc) *Handler {
	return &Handler{quotes: quotes, generateTokenFn: g, hashFn: f}
}

const defaultClientAddr = "localhost"

type Handler struct {
	quotes          *Quotes
	hashFn          HashFunc
	generateTokenFn GenerateTokenFunc
}

func (h *Handler) ReqChallenge(r proto.Request, w *proto.ResponseWriter) {
	ctx := r.Context()
	logger := logging.FromContext(ctx).Named("ReqChallenge")
	addr, ok := ctx.Value(server.ClientSubjectCtxKey).(server.ContextSubjectKey)
	if !ok {
		addr = defaultClientAddr
	}

	challenge, err := h.quotes.MakeChallenge(string(addr))
	if err != nil {
		logger.Errorf("MakeChallenge: %v", err)
		if _, err = w.SendErr(server.ErrInternalServer.Error()); err != nil {
			logger.Errorf("send err: %v", err)

			return
		}
	}

	b, cancel := GetBuilder()
	defer cancel()

	_, _ = fmt.Fprintf(b, "%s\n%s", h.generateTokenFn(h.hashFn, challenge), challenge.String())

	if _, err = w.SendRSV(b.String()); err != nil {
		logger.Errorf("send rsv: %v", err)
	}
}

func (h *Handler) GetResource(r proto.Request, w *proto.ResponseWriter) {
	ctx := r.Context()
	logger := logging.FromContext(ctx).Named("GetResource")

	resource, err := h.quotes.GetResource()
	if err != nil {
		logger.Errorf("GetResource: %v", err)
		if _, err = w.SendErr(server.ErrInternalServer.Error()); err != nil {
			logger.Errorf("send err: %v", err)
			return
		}
	}

	if _, err = w.SendRSV(resource); err != nil {
		logger.Errorf("send rsv: %v", err)
		return
	}
}
