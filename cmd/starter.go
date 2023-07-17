package cmd

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"balance_checker/config"
	"balance_checker/database"
	"balance_checker/worker"
)

func Start() error {
	c, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("get config err: %s", err)
	}

	dbc, db, err := database.GetController(c)
	if err != nil {
		return fmt.Errorf("connect to db err: %s", err)
	}

	defer func() {
		_ = db.Close()
	}()

	workerStart(c, dbc)

	return nil
}

func workerStart(conf *config.Config, dbc *database.Controller) {
	user := make(chan database.User)
	err := make(chan error)
	urls := make(chan string)
	proxies := make(chan string, 1)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func(proxies chan string, listProxies []string, wg *sync.WaitGroup) {
		defer wg.Done()

		for i := 0; i < len(listProxies) && i < conf.NumProxies; i++ {
			proxies <- listProxies[i]
		}

		close(proxies)
	}(proxies, conf.Proxies, &wg)

	wg.Add(1)
	go func(urls chan string, listUrls []string, wg *sync.WaitGroup, proxies chan string) {
		defer wg.Done()

		for _, u := range listUrls {
			urls <- u
		}

		close(urls)
	}(urls, conf.URLs, &wg, proxies)

	go func() {
		for {
			select {
			case u := <-user:
				if u.Balance > 0 {
					fmt.Println(u.Id, u.Balance)
					if err := dbc.AddBalance(u.Id, u.Balance); err != nil {
						log.Printf("add balance err: %s", err)
					}
				}
			case e := <-err:
				if !strings.Contains(e.Error(), "urls are over") {
					log.Printf("worker err: %s", e)
				}
			}
		}
	}()

	for proxy := range proxies {
		go func(proxy string, urls chan string, user chan database.User, err chan error, wg *sync.WaitGroup) {
			c := worker.GetNewWorker(conf, wg, proxy, urls, user, err)

			c.StartNew()
		}(proxy, urls, user, err, &wg)
	}

	wg.Wait()
}
