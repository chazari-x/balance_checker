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

	"balance_checker/config"
	"balance_checker/database"
	"github.com/PuerkitoBio/goquery"
)

type Worker interface {
	StartNew()
}

type Controller struct {
	c     *config.Config
	wg    *sync.WaitGroup
	proxy string
	urls  chan string
	user  chan database.User
	err   chan error
}

func GetNewWorker(c *config.Config, wg *sync.WaitGroup, proxy string, url chan string, u chan database.User, err chan error) *Controller {
	return &Controller{c: c, wg: wg, proxy: proxy, urls: url, user: u, err: err}
}

func (c *Controller) StartNew() {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for url := range c.urls {
			fmt.Println(url)

			resp, err := http.Get(url)
			if err != nil {
				c.err <- err
				return
			}

			defer func() {
				_ = resp.Body.Close()
			}()

			// Загружаем HTML документ с помощью goquery
			doc, err := goquery.NewDocumentFromReader(resp.Body)
			if err != nil {
				c.err <- err
				return
			}

			element := doc.Find(`body > main > section:nth-child(2) > 
				div.-mx-4.w-\[calc\(100\%\+32px\)\].overflow-x-auto.sm\:mx-0.sm\:w-full.rounded-lg.bg-gray-800.pb-4 > 
				table > tbody > tr > td:nth-child(2) > div > span.mt-2.text-xxs.text-zinc-500`)
			if element.Length() <= 0 {
				c.err <- errors.New("element not found")
				return
			}

			strBalance := regexp.MustCompile(`[0-9]+,[0-9]+.[0-9]+`).FindString(element.Text())

			balance, err := strconv.ParseFloat(strings.ReplaceAll(strBalance, ",", ""), 64)
			if err != nil {
				c.err <- err
				return
			}

			c.user <- database.User{
				User:    regexp.MustCompile(`users/\S+`).FindString(url)[6:],
				Balance: balance,
			}

			time.Sleep(time.Second)
		}

		c.err <- fmt.Errorf("urls is nil")
	}()
}
