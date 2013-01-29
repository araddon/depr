package main

import (
	"github.com/araddon/gou"
	"log"
	"os"
)

func main() {
	gou.SetLogger(log.New(os.Stderr, "", log.Ltime|log.Lshortfile), "debug")
	gou.Debug("hello")
}
