package main

import (
	_ "embed"
	"log"
	"os"
	"runtime/debug"

	"github.com/hugorosario/trimuitools/gui"
)

func main() {
	defer func() {
		gui.FreeGUI()
		if r := recover(); r != nil {
			log.Printf("Unhandled error: %v\n", r)
			log.Println("Stack trace:")
			debug.PrintStack()
			os.Exit(-1)
		}
	}()
	if err := gui.InitGUI(); err != nil {
		panic(err)
	}

	gui.Show()
}
