package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/robotomize/powwy/internal/client"
	"github.com/robotomize/powwy/internal/shutdown"
	"github.com/robotomize/powwy/pkg/hashcash"
	"github.com/spf13/cobra"
)

var (
	maxWorkersNum int
	maxIterations int
	maxDuration   time.Duration
	addr          string
	network       string
	dos           bool
)

var rootCmd = &cobra.Command{
	Use:   "",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := shutdown.New()
		defer cancel()

		logger := log.New(os.Stdout, "cli: ", log.Lmsgprefix)

		opts := []hashcash.PoolOption{hashcash.WithWorkerNum(maxWorkersNum)}

		if maxDuration > -1 {
			opts = append(opts, hashcash.WithPoolDuration(maxDuration))
		}

		if maxIterations > -1 {
			opts = append(opts, hashcash.WithPoolMaxIterations(maxIterations))
		}

		logger.Printf("try connect to %s...", addr)

		c := client.NewClient(
			client.Config{
				Addr:    addr,
				Network: network,
			},
		)

		if dos {
			for {
				msg, info, err := try(ctx, c, opts...)
				if err != nil {
					logger.Fatal(err)
				}

				logger.Printf("ts: %s, hash: %s, msg: %s", info.Time, info.Header.Hash(), msg)
				select {
				case <-ctx.Done():
					return nil
				default:
				}
			}
		}

		msg, info, err := try(ctx, c, opts...)
		if err != nil {
			logger.Fatal(err)
		}

		logger.Printf("ts: %s, hash: %s, msg: %s", info.Time, info.Header.Hash(), msg)

		return nil
	},
}

func try(ctx context.Context, c *client.Client, opts ...hashcash.PoolOption) (string, hashcash.PoolInfo, error) {
	header, err := c.SendREQ(ctx)
	if err != nil {
		return "", hashcash.PoolInfo{}, err
	}

	info, err := hashcash.ComputeWithPool(ctx, header, opts...)
	if err != nil {
		return "", hashcash.PoolInfo{}, err
	}

	text, err := c.SendRES(ctx, info.Header.String())
	if err != nil {
		return "", hashcash.PoolInfo{}, err
	}

	return string(text), info, nil
}

func init() {
	rootCmd.PersistentFlags().IntVarP(&maxWorkersNum, "workers", "w", 2, "-w 4")
	rootCmd.PersistentFlags().IntVarP(&maxIterations, "iterations", "i", -1, "-i 1000000")
	rootCmd.PersistentFlags().DurationVarP(&maxDuration, "duration", "d", -1, "-d 10s")
	rootCmd.PersistentFlags().StringVarP(&addr, "addr", "a", "localhost:3333", "-a localhost:3333")
	rootCmd.PersistentFlags().StringVarP(&network, "network", "n", "tcp", "-n tcp4")
	rootCmd.PersistentFlags().BoolVarP(&dos, "dos", "s", false, "-s true")
}
