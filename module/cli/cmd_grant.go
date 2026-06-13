package cli

import (
	"fmt"

	"github.com/gwaylib/errors"
	"github.com/gwaysys/bcstorage/module/client"
	"github.com/urfave/cli/v2"
)

var GrantCmd = &cli.Command{
	Name:  "grant",
	Usage: "[remote path]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "kind",
			Usage: "'a' for readonly, and not expired",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "token",
			Usage: "specified value, using random value if not set",
			Value: "",
		},
		&cli.IntFlag{
			Name:  "exp",
			Usage: "auth expire after exp minutes",
			Value: 60,
		},
	},
	Action: func(cctx *cli.Context) error {
		if !cctx.Args().Present() {
			return fmt.Errorf("arguments with [remote path]")
		}
		args := cctx.Args()
		if args.Len() != 1 {
			return fmt.Errorf("arguments with [remote path]")
		}
		remotePath := args.Get(0)

		// TODO: process the download
		ctx := cctx.Context
		authFile := remotePath
		ac := client.NewAuthClient(_authApiFlag,
			cctx.String("user"),
			cctx.String("passwd"),
		)
		newToken, err := ac.NewFileToken(ctx, authFile, cctx.String("kind"), cctx.String("token"), cctx.Int("exp"))
		if err != nil {
			return errors.As(err)
		}
		fmt.Printf("grant:%s, token:%s\n", cctx.String("kind"), newToken)
		return nil
	},
}
