package output

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	Dir  string `yaml:"dir"`
	File string `yaml:"file"`
}

type Output struct {
	outputFile *os.File
	config     Config
}

func NewStore(cfg Config) (*Output, error) {
	newFile := fmt.Sprintf("%s/%d-%s", cfg.Dir, time.Now().Unix(), cfg.File)
	_, err := os.Stat(newFile)
	if os.IsNotExist(err) {
		_, err = os.Create(newFile)
		if err != nil {
			return nil, err
		}
	}

	// Открытие файла для записи
	file, err := os.OpenFile(newFile, os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	o := Output{
		outputFile: file,
		config:     cfg,
	}

	return &o, nil
}

func (o *Output) Write(output string) error {
	// Запись строк в файл построчно
	_, err := fmt.Fprintln(o.outputFile, output)
	if err != nil {
		return err
	}

	return nil
}
