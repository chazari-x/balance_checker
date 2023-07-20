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
	LOOP:
		var timeouts int

		timeout := time.Duration(c.c.TimeOut) * time.Second
		transport := c.proxyStore.Get()
		transport.TLSHandshakeTimeout = timeout
		client := &http.Client{Transport: transport, Timeout: timeout}

	OUT:
		for {
			select {
			case u := <-c.inputStore.Items:
				req, err := http.NewRequest("GET", fmt.Sprintf("https://dydx.l2beat.com/users/%s", u.Value), nil)
				if err != nil {
					c.err <- fmt.Errorf("http request err: %s", err)
					go func() {
						c.inputStore.Items <- u
					}()
					goto LOOP
				}

				resp, err := client.Do(req)
				if err != nil {
					if !strings.Contains(err.Error(), "context deadline exceeded") {
						c.err <- fmt.Errorf("client do err: %s", err)
						go func() {
							c.inputStore.Items <- u
						}()
						goto LOOP
					}

					go func() {
						c.inputStore.Items <- u
					}()

					timeouts++
					if timeouts > 10 {
						c.err <- fmt.Errorf("client do err: %s %d: worker closed", err, timeouts)
						goto LOOP
					}

					c.err <- fmt.Errorf("client do: %s %d", err, timeouts)
					continue
				}

				doc, err := goquery.NewDocumentFromReader(resp.Body)
				if err != nil {
					c.err <- fmt.Errorf("goquery new document from reader err: %s", err)
					_ = resp.Body.Close()
					continue
				}

				_ = resp.Body.Close()

				element := doc.Find(`body > main > section:nth-child(2) > 
				div.-mx-4.w-\[calc\(100\%\+32px\)\].overflow-x-auto.sm\:mx-0.sm\:w-full.rounded-lg.bg-gray-800.pb-4 > 
				table > tbody > tr > td:nth-child(2) > div > span.mt-2.text-xxs.text-zinc-500`)
				if element.Length() <= 0 {
					c.err <- errors.New("element not found")
					continue
				}

				strBalance := regexp.MustCompile(`[0-9]+,[0-9]+.[0-9]+`).FindString(element.Text())

				balance, err := strconv.ParseFloat(strings.ReplaceAll(strBalance, ",", ""), 64)
				if err != nil {
					c.err <- fmt.Errorf("strconv parse float err: %s", err)
					continue
				}

				if balance > 0 {
					c.user <- User{
						Id:      u.Value,
						Balance: balance,
					}
				}
			case <-time.After(time.Second * 30):
				break OUT
			}

		}

		c.err <- fmt.Errorf("urls are over")
	}()
}
