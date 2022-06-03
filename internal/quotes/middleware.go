package quotes

import (
	"bytes"
	"errors"
	"hash"
	"time"

	"github.com/robotomize/powwy/internal/logging"
	"github.com/robotomize/powwy/internal/server"
	"github.com/robotomize/powwy/pkg/hashcash"
	"github.com/robotomize/powwy/pkg/proto"
)

var (
	ErrUnknownNonce     = errors.New("nonce not found")
	ErrHeaderValidation = errors.New("header invalid")
	ErrHashWrong        = errors.New("hash wrong")
	ErrTokenWrong       = errors.New("token wrong")
	ErrHeaderExpired    = errors.New("header expired")
)

func PoWMiddleware(next server.HandleFunc, hashFn func() hash.Hash, generateTokenFn GenerateTokenFunc) server.HandleFunc {
	return func(r proto.Request, w *proto.ResponseWriter) {
		ctx := r.Context()
		logger := logging.FromContext(ctx).Named("PoWMiddleware")

		contents := bytes.Split(r.Body, []byte("\n"))
		if len(contents) < 2 {
			if _, err := w.SendErr(ErrUnknownNonce.Error()); err != nil {
				logger.Errorf("send err: %v", err)
			}

			return
		}

		header, err := hashcash.Parse(string(contents[1]))
		if err != nil {
			if _, err = w.SendErr(ErrHeaderValidation.Error()); err != nil {
				logger.Errorf("send err: %v", err)
			}

			return
		}

		if string(contents[0]) != generateTokenFn(hashFn, header) {
			if _, err = w.SendErr(ErrTokenWrong.Error()); err != nil {
				logger.Errorf("send err: %v", err)
			}

			return
		}

		t := time.Unix(header.ExpiredAt, 0)
		if !time.Now().Before(t) {
			if _, err = w.SendErr(ErrHeaderExpired.Error()); err != nil {
				logger.Errorf("send err: %v", err)
			}

			return
		}

		if !header.IsValid() {
			if _, err = w.SendErr(ErrHashWrong.Error()); err != nil {
				logger.Errorf("send err: %v", err)
				return
			}
		}

		next(r, w)
	}
}
