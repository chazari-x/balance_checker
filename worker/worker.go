package worker

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"balance_checker/cmd"
	"balance_checker/database"
	"github.com/PuerkitoBio/goquery"
)

type Worker interface {
	StartNew()
}

type Controller struct {
	c     *cmd.Config
	wg    *sync.WaitGroup
	proxy string
	urls  chan string
	user  chan database.User
	err   chan error
}

func NewWorker(c *cmd.Config, wg *sync.WaitGroup, proxy string, url chan string, u chan database.User, err chan error) *Controller {
	return &Controller{c: c, wg: wg, proxy: proxy, urls: url, user: u, err: err}
}

func (c *Controller) StartNew() {
	c.wg.Add(1)
	go func() {
		var timeouts int

		defer c.wg.Done()
		for u := range c.urls {
			var (
				proxyURLParsed, _ = url.Parse(c.proxy)
				proxy             = http.ProxyURL(proxyURLParsed)
				timeout           = time.Duration(c.c.TimeOut) * time.Second
				transport         = &http.Transport{
					Proxy:               proxy,
					TLSHandshakeTimeout: timeout,
				}
				client = &http.Client{Transport: transport, Timeout: timeout}
			)

			req, err := http.NewRequest("GET", u, nil)
			if err != nil {
				c.err <- fmt.Errorf("http request err: %s", err)
				go func() {
					c.urls <- u
				}()
				return
			}

			resp, err := client.Do(req)
			if err != nil {
				if !strings.Contains(err.Error(), "context deadline exceeded") {
					c.err <- fmt.Errorf("client do err: %s", err)
					go func() {
						c.urls <- u
					}()
					return
				}

				go func() {
					c.urls <- u
				}()

				timeouts++
				if timeouts > 10 {
					c.err <- fmt.Errorf("client do err: %s %d: worker closed", err, timeouts)
					return
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

			c.user <- database.User{
				Id:      regexp.MustCompile(`users/\S+`).FindString(u)[6:],
				Balance: balance,
			}
		}

		c.err <- fmt.Errorf("urls are over")
	}()
}
