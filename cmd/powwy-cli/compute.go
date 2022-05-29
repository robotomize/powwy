package main

import (
	"context"
	"log"
	"os"

	"github.com/robotomize/powwy/pkg/hashcash"
	"github.com/spf13/cobra"
)

var computeCmd = &cobra.Command{
	Use:   "compute",
	Short: "Usage: powwy-cli compute <header> <header 2> <header n>",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := log.New(os.Stdout, "cli: ", log.Lmsgprefix)
		if len(args) == 0 {
			logger.Print(cmd.Short)
		}

		opts := []hashcash.PoolOption{hashcash.WithWorkerNum(maxWorkersNum)}

		if maxDuration > -1 {
			opts = append(opts, hashcash.WithPoolDuration(maxDuration))
		}

		if maxIterations > -1 {
			opts = append(opts, hashcash.WithPoolMaxIterations(maxIterations))
		}

		for _, arg := range args {
			header, err := hashcash.Parse(arg)
			if err != nil {
				logger.Printf("error: header %s invalid, "+
					"user header format: 1:20:1665396610:localhost:sha-512:hVscDCMZcS1WYg==:BQAAAAAAAAA=\n", arg)
				continue
			}

			logger.Printf("try compute hash for %s", arg)

			pool, err := hashcash.ComputeWithPool(context.Background(), header, opts...)
			if err != nil {
				logger.Printf("error: %v\n", err)
				continue
			}

			logger.Println("solution found:")
			logger.Printf("hash: %s", pool.Header.Hash())
			logger.Printf("header: %s", pool.Header.String())
			logger.Printf("ts: %s\n------", pool.Time)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(computeCmd)
}
