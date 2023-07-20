package proxy

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
)

type Config struct {
	File string `yaml:"file"`
}

type Proxy struct {
	Value string `json:"value"`
}

type Store struct {
	items  []Proxy
	config Config
	m      sync.Mutex
	i      int
}

func NewStore(cfg Config) (*Store, error) {
	s := Store{
		config: cfg,
	}

	err := s.upload()
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (s *Store) Get() *http.Transport {
	s.m.Lock()

	if s.i > len(s.items) {
		s.i = 0
	}

	pr := s.items[s.i].Value
	s.m.Unlock()

	proxyURLParsed, err := url.Parse(fmt.Sprintf("http://%s", pr))
	if err != nil {
		return s.Get()
	}
	proxy := http.ProxyURL(proxyURLParsed)
	transport := &http.Transport{
		Proxy:               proxy,
		TLSHandshakeTimeout: 0,
	}

	return transport
}

func (s *Store) upload() error {
	if s.config.File != "" {
		return s.uploadFromFile()
	}

	return errors.New("proxy config err")
}

func (s *Store) uploadFromFile() error {
	f, err := os.Open(s.config.File)
	if err != nil {
		return fmt.Errorf("open proxy file: %v", err)
	}

	defer f.Close()

	proxyList := make([]Proxy, 0)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		proxy := Proxy{Value: scanner.Text()}
		proxyList = append(proxyList, proxy)
	}

	s.items = proxyList

	return nil
}
