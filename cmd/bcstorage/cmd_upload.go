package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/gwaysys/bcstorage/module/client"
)

var uploadCmd = &cli.Command{
	Name:  "upload",
	Usage: "[local path] [remote path]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "mode",
			Usage: "downlad mode, support mode: 'http', 'TODO:tcp'",
			Value: "http",
		},
		&cli.StringFlag{
			Name:  "space",
			Usage: "user spacename, same as username when not set",
			Value: "",
		},
	},
	Action: func(cctx *cli.Context) error {
		if !cctx.Args().Present() {
			return fmt.Errorf("arguments with [local path] [remote path]")
		}
		args := cctx.Args()
		if args.Len() != 2 {
			return fmt.Errorf("arguments with [local path] [remote path]")
		}
		localPath := args.Get(0)
		remotePath := args.Get(1)

		end := make(chan os.Signal, 2)
		switch cctx.String("mode") {
		case "http":
			go func() {
				// TODO: process the download
				ctx := cctx.Context
				authFile := remotePath
				ac := client.NewAuthClient(_authApiFlag,
					cctx.String("user"),
					cctx.String("passwd"),
				)
				space := cctx.String("space")
				if len(space) == 0 {
					space = cctx.String("user")
				}
				newToken, err := ac.NewFileToken(ctx, space, authFile)
				if err != nil {
					panic(err)
				}
				log.Infof("start upload: %s->%s", localPath, remotePath)
				startTime := time.Now()
				fc := client.NewHttpClient(_httpApiFlag, space, authFile, string(newToken))
				if err := fc.Upload(ctx, localPath, remotePath); err != nil {
					panic(err)
				}
				log.Infof("end upload: %s->%s, took:%s", localPath, remotePath, time.Now().Sub(startTime))
				end <- os.Kill
			}()
		default:
			return fmt.Errorf("unknow mode '%s'", cctx.String("mode"))

		}
		// TODO: show the process
		signal.Notify(end, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT)
		<-end
		return nil
	},
}
