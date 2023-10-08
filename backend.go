package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

type Backend interface {
	AddFlags()
	PublishStats() error
	Activate()
	Close()
}

type TextBackend struct {
	Filename   string
	Separator  string
	IsActive   bool
	FileHandle *os.File
	Prefix     string
	Suffix     string
}

func (t *TextBackend) AddFlags() {
	OsEnvWithDefault("RLM_TEXT_FILE", "stdout")
	OsEnvWithDefault("RLM_TEXT_SEPARATOR", "\t")
	OsEnvWithDefault("RLM_TEXT_PREFIX", "")
	OsEnvWithDefault("RLM_TEXT_SUFFIX", "")

	flag.BoolVar(&t.IsActive, "text", false, "Activate text backend")
	flag.StringVar(&t.Filename, "text-filename", "stdout", "Filename to publish stats to")
	flag.StringVar(&t.Separator, "text-separator", "\t", "Separator to use between fields")
	flag.StringVar(&t.Prefix, "text-prefix", "", "Prefix to prepend to each stat")
	flag.StringVar(&t.Suffix, "text-suffix", "", "Suffix to append to each stat")
}

func (t *TextBackend) Activate() {
	if !t.IsActive {
		return
	}
	if t.Filename == "stdout" {
		t.FileHandle = os.Stdout
		return
	}
	fh, err := os.OpenFile(t.Filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error initializing file: %v", err)
	}
	t.FileHandle = fh
}

func (t *TextBackend) Close() {
	if !t.IsActive {
		return
	}
	if t.Filename == "stdout" {
		return
	}
	t.FileHandle.Close()
}

func (t *TextBackend) PublishStats() error {
	if !t.IsActive {
		return nil
	}
	for _, stat := range metricMemory.Metrics {
		_, err := fmt.Fprintf(t.FileHandle, "%v%v%v%v%v\n", t.Prefix, stat.Name, t.Suffix, t.Separator, stat.Value)
		if err != nil {
			return err
		}
	}
	t.FileHandle.Sync()
	return nil
}
