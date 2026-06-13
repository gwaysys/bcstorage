package main

import (
	"os"

	"github.com/gwaylib/log"
	bcli "github.com/gwaysys/bcstorage/module/cli"
)

func main() {
	app := bcli.NewApp()
	app.Setup()
	if err := app.Run(os.Args); err != nil {
		log.Warnf("%+v", err)
		return
	}
}
