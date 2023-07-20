package input

import (
	"bufio"
	"errors"
	"fmt"
	"log"
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
		return fmt.Errorf("open input file: %v", err)
	}

	go func(f *os.File) {
		defer func() {
			_ = f.Close()
		}()

		scanner := bufio.NewScanner(f)
		i := 0
		for scanner.Scan() {
			i++
			if i%1000 == 0 {
				log.Printf("current line %d", i)
			}
			u := URL{Value: scanner.Text()}
			s.Items <- u
		}
	}(f)

	return nil
}
