package output

import (
	"fmt"
	"log"
	"os"
)

type Config struct {
	File string `yaml:"file"`
}

type Output struct {
	outputFile *os.File
	config     Config
}

func NewStore(cfg Config) (*Output, error) {
	_, err := os.Stat(cfg.File)
	if os.IsNotExist(err) {
		_, err = os.Create(cfg.File)
		if err != nil {
			return nil, err
		}
	}

	// Открытие файла для записи
	file, err := os.OpenFile(cfg.File, os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	o := Output{
		outputFile: file,
		config:     cfg,
	}

	return &o, nil
}

func (o *Output) Write(output string) error {
	// Запись строк в файл построчно
	log.Print("output")
	_, err := fmt.Fprintln(o.outputFile, output)
	if err != nil {
		return err
	}

	return nil
}
