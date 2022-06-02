package quotes

import (
	"errors"

	"github.com/robotomize/powwy/internal/logging"
	"github.com/robotomize/powwy/internal/server"
	"github.com/robotomize/powwy/pkg/hashcash"
	"github.com/robotomize/powwy/pkg/proto"
)

var (
	ErrUnknownNonce     = errors.New("nonce not found or header expired")
	ErrHeaderValidation = errors.New("header invalid")
	ErrHashWrong        = errors.New("hash wrong")
	ErrInternalServer   = errors.New("internal server")
)

func PoWMiddleware(next server.HandleFunc, s Store) server.HandleFunc {
	return func(r proto.Request, w *proto.ResponseWriter) {
		ctx := r.Context()
		logger := logging.FromContext(ctx).Named("PoWMiddleware")

		header, err := hashcash.Parse(string(r.Body))
		if err != nil {
			if _, err = w.SendErr(ErrHeaderValidation.Error()); err != nil {
				logger.Errorf("send err: %v", err)
			}

			return
		}

		originHeader, ok := s.Lookup(header.Nonce)
		if !ok {
			if _, err = w.SendErr(ErrUnknownNonce.Error()); err != nil {
				logger.Errorf("send err: %v", err)
			}

			return
		}

		if originHeader.Alg != header.Alg {
			if _, err = w.SendErr(ErrHeaderValidation.Error()); err != nil {
				logger.Errorf("send err: %v", err)
			}

			return
		}

		if originHeader.Difficult != header.Difficult {
			if _, err = w.SendErr(ErrHeaderValidation.Error()); err != nil {
				logger.Errorf("send err: %v", err)
			}

			return
		}

		if originHeader.Version != header.Version {
			if _, err = w.SendErr(ErrHeaderValidation.Error()); err != nil {
				logger.Errorf("send err: %v", err)
			}

			return
		}

		if originHeader.Subject != header.Subject {
			if _, err = w.SendErr(ErrHeaderValidation.Error()); err != nil {
				logger.Errorf("send err: %v", err)
			}

			return
		}

		if originHeader.ExpiredAt != header.ExpiredAt {
			if _, err = w.SendErr(ErrHeaderValidation.Error()); err != nil {
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

		if err = s.Delete(header.Nonce); err != nil {
			if _, err = w.SendErr(ErrInternalServer.Error()); err != nil {
				logger.Errorf("send err: %v", err)
			}

			return
		}

		next(r, w)
	}
}
