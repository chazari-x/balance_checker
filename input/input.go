package input

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"sync"
)

type Config struct {
	File string `yaml:"file"`
}

type URL struct {
	Value string `json:"value"`
}

type Store struct {
	Items  chan URL
	config Config
	m      sync.Mutex
}

func NewStore(cfg Config) (*Store, error) {
	s := Store{
		Items:  make(chan URL),
		config: cfg,
	}

	err := s.upload()
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (s *Store) upload() error {
	if s.config.File == "" {
		return errors.New("input config err")
	}

	f, err := os.Open(s.config.File)
	if err != nil {
		return fmt.Errorf("open proxy file: %v", err)
	}

	go func(f *os.File) {
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			u := URL{Value: scanner.Text()}
			s.Items <- u
		}
	}(f)

	return nil
}
