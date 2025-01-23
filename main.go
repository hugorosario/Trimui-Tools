package main

import (
	_ "embed"
	"log"
	"os"
	"runtime/debug"

	"github.com/hugorosario/trimuitools/app"
)

func main() {
	defer func() {
		app.FreeGUI()
		if r := recover(); r != nil {
			log.Printf("Unhandled error: %v\n", r)
			log.Println("Stack trace:")
			debug.PrintStack()
			os.Exit(-1)
		}
	}()

	if err := app.InitGUI(); err != nil {
		panic(err)
	}
	app.Show()
}
