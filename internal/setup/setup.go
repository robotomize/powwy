package setup

import (
	"context"
	"crypto/md5" //nolint
	"fmt"
	"hash"
	"net"

	"github.com/robotomize/powwy/internal/quotes"
	"github.com/robotomize/powwy/internal/server"
	"github.com/robotomize/powwy/pkg/proto"
	"github.com/sethvargo/go-envconfig"
)

type Environment struct {
	Server *server.Server
}

func Setup(ctx context.Context) (Environment, error) {
	var env Environment
	var config Config
	if err := envconfig.Process(ctx, &config); err != nil {
		return env, fmt.Errorf("env processing: %w", err)
	}

	handler := quotes.NewHandler(
		quotes.NewQuotes(config.Quotes), func() hash.Hash {
			return md5.New()
		}, quotes.GenerateToken,
	)

	l, err := net.Listen(config.Server.Network, config.Server.Addr)
	if err != nil {
		return env, fmt.Errorf("net.Listen: %w", err)
	}

	srv, err := server.New(l, config.Server.GracefulConnCloseDeadline)
	if err != nil {
		return env, fmt.Errorf("server.New: %w", err)
	}

	srv.HandleFunc(proto.REQ, handler.ReqChallenge)
	srv.HandleFunc(
		proto.RES, quotes.PoWMiddleware(handler.GetResource, func() hash.Hash {
			return md5.New()
		}, quotes.GenerateToken),
	)

	env.Server = srv

	return env, nil
}
