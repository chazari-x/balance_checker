package cmd

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"balance_checker/input"
	"balance_checker/output"
	"balance_checker/proxy"
	"balance_checker/worker"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "worker",
		Short: "Worker",
		Long:  "Worker",
		Run: func(cmd *cobra.Command, args []string) {
			log.Println("starting..")

			ctx := cmd.Context()
			cfg := configFromContext(ctx)

			proxyStore, err := proxy.NewStore(cfg.ProxyConfig)
			if err != nil {
				log.Fatal(err)
			}

			inputStore, err := input.NewStore(cfg.InputConfig)
			if err != nil {
				log.Fatal(err)
			}

			outputStore, err := output.NewStore(cfg.OutputConfig)
			if err != nil {
				log.Fatal(err)
			}

			userCh := make(chan worker.User)
			errCh := make(chan error)

			wg := sync.WaitGroup{}

			newWorker := worker.NewWorker(&cfg.WorkerConfig, &wg, userCh, errCh, proxyStore, inputStore)

			go func() {
				for {
					select {
					case u := <-userCh:
						if err := outputStore.Write(fmt.Sprintf("%s, %.2f", u.Id, u.Balance)); err != nil {
							log.Printf("write err: %s", err)
						}
					case e := <-errCh:
						if !strings.Contains(e.Error(), "urls are over") {
							log.Printf("worker err: %s", e)
						}
					}
				}
			}()

			for i := 0; i < cfg.WorkerConfig.Threads; i++ {
				newWorker.Start()
			}

			wg.Wait()
		},
	})
}
