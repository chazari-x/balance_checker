package worker

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"balance_checker/input"
	"balance_checker/proxy"
	"github.com/PuerkitoBio/goquery"
)

type Controller struct {
	c          *Config
	wg         *sync.WaitGroup
	user       chan User
	err        chan error
	proxyStore *proxy.Store
	inputStore *input.Store
}

type Config struct {
	Threads int `yaml:"threads"`
	TimeOut int `yaml:"timeout"`
}

type User struct {
	Id      string  `json:"id"`
	Balance float64 `json:"balance"`
}

func NewWorker(c *Config, wg *sync.WaitGroup, u chan User, err chan error, proxyStore *proxy.Store, inputStore *input.Store) *Controller {
	return &Controller{c: c, wg: wg, user: u, err: err, proxyStore: proxyStore, inputStore: inputStore}
}

func (c *Controller) Start() {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		timeout := time.Duration(c.c.TimeOut) * time.Second

	LOOP:
		var timeouts int

		transport := c.proxyStore.Get()
		transport.TLSHandshakeTimeout = timeout
		client := &http.Client{Transport: transport, Timeout: timeout}

	OUT:
		for {
			select {
			case u := <-c.inputStore.Items:
				req, err := http.NewRequest("GET", fmt.Sprintf("https://dydx.l2beat.com/search?query=%s", u.Value), nil)
				if err != nil {
					go func() {
						c.err <- fmt.Errorf("http request err: %s", err)
					}()
					go func() {
						c.inputStore.Items <- u
					}()
					goto LOOP
				}

				resp, err := client.Do(req)
				if err != nil {
					if !strings.Contains(err.Error(), "context deadline exceeded") {
						go func() {
							c.err <- fmt.Errorf("client do err: %s. Changing the proxy", err)
						}()
						go func() {
							c.inputStore.Items <- u
						}()
						goto LOOP
					}

					go func() {
						c.inputStore.Items <- u
					}()

					timeouts++
					if timeouts >= 10 {
						go func() {
							c.err <- fmt.Errorf("client do err: %s. %d: changing the proxy", err, timeouts)
						}()
						goto LOOP
					}

					continue
				}

				if resp.StatusCode != 200 {
					go func() {
						c.err <- fmt.Errorf(resp.Status)
					}()

					break OUT
				}

				doc, err := goquery.NewDocumentFromReader(resp.Body)
				if err != nil {
					if !strings.Contains(err.Error(), "context deadline exceeded") {
						go func() {
							c.err <- fmt.Errorf("goquery new document from reader err: %s", err)
						}()
						_ = resp.Body.Close()
						go func() {
							c.inputStore.Items <- u
						}()
						continue
					}

					go func() {
						c.inputStore.Items <- u
					}()

					timeouts++
					if timeouts >= 10 {
						go func() {
							c.err <- fmt.Errorf("client do err: %s. %d: changing the proxy", err, timeouts)
						}()
						goto LOOP
					}

					continue
				}

				_ = resp.Body.Close()

				selector := `body > main > section:nth-child(2) > 
					div.-mx-4.w-\[calc\(100\%%\+32px\)\].overflow-x-auto.sm\:mx-0.sm\:w-full.rounded-lg.bg-gray-800.pb-4 > 
					table > tbody > tr:nth-child(%d) > td:nth-child(2) > div > span.mt-2.text-xxs.text-zinc-500`
				var allBalance float64
				for i := 1; ; i++ {
					element := doc.Find(fmt.Sprintf(selector, i))
					if element.Length() <= 0 {
						if i == 1 {
							go func() {
								c.err <- errors.New("element not found")
							}()
							continue OUT
						} else {
							break
						}
					}

					strBalance := regexp.MustCompile(`-?\$[0-9,]+.?[0-9]{0,2}`).FindString(element.Text())

					balance, err := strconv.ParseFloat(strings.ReplaceAll(strings.ReplaceAll(strBalance, "$", ""), ",", ""), 64)
					if err != nil {
						go func() {
							c.err <- fmt.Errorf("strconv parse float err: %s", err)
						}()
						continue
					}

					allBalance += balance
				}

				if allBalance > 0 {
					c.user <- User{
						Id:      u.Value,
						Balance: allBalance,
					}
				}
			case <-time.After(time.Second * 15):
				break OUT
			}

		}

		c.err <- fmt.Errorf("urls are over")
	}()
}
