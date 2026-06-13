package cli

import (
	"fmt"

	"github.com/gwaylib/errors"
	"github.com/gwaysys/bcstorage/module/client"
	"github.com/gwaysys/bcstorage/module/server"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"
)

var SysCmd = &cli.Command{
	Name: "sys",
	Subcommands: []*cli.Command{
		StatusCmd,
		AuthNewCmd,
		AuthResetCmd,
	},
}
var StatusCmd = &cli.Command{
	Name:  "status",
	Usage: "status of storage",
	Action: func(cctx *cli.Context) error {
		// TODO: process the download
		ctx := cctx.Context
		ac := client.NewAuthClient(
			_authApiFlag,
			cctx.String("user"),
			cctx.String("passwd"),
		)
		result, err := ac.Check(ctx)
		if err != nil {
			return errors.As(err)
		}
		fmt.Printf("%s", string(result))
		return nil
	},
}
var AuthNewCmd = &cli.Command{
	Name:  "adduser",
	Usage: "adduser [username]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "space",
			Usage: "user space name, empty same with username",
			Value: "",
		},
	},
	Action: func(cctx *cli.Context) error {
		if !cctx.Args().Present() {
			return fmt.Errorf("arguments with [username]")
		}
		args := cctx.Args()
		if args.Len() != 1 {
			return fmt.Errorf("arguments with [username]")
		}
		user := args.Get(0)
		space := cctx.String("space")
		if len(space) == 0 {
			space = user
		}

		// TODO: process the download
		ctx := cctx.Context
		ac := client.NewAuthClient(_authApiFlag,
			cctx.String("user"),
			cctx.String("passwd"),
		)
		passwd, err := ac.AddUser(ctx, user, space)
		if err != nil {
			return errors.As(err)
		}
		fmt.Printf("password: %s\n", string(passwd))
		return nil
	},
}
var AuthResetCmd = &cli.Command{
	Name:  "reset-passwd",
	Usage: "reset user passwd",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "local",
			Value: false,
			Usage: "reset password to leveldb directly after stop the daemon",
		},
		&cli.StringFlag{
			Name:  "local-cfg-path",
			Value: "",
			Usage: "cfg-path in local mode",
		},
		&cli.StringFlag{
			Name:  "new-passwd",
			Value: "",
			Usage: "new password",
		},
	},
	Action: func(cctx *cli.Context) error {
		if !cctx.Args().Present() {
			return fmt.Errorf("arguments with [username]")
		}
		args := cctx.Args()
		if args.Len() != 1 {
			return fmt.Errorf("arguments with [username]")
		}
		user := args.Get(0)

		// TODO: process the download
		ctx := cctx.Context
		if cctx.Bool("local") {
			localCfgPath, err := homedir.Expand(cctx.String("local-cfg-path"))
			if err != nil {
				return errors.As(err)
			}
			if len(localCfgPath) == 0 {
				return errors.New("local-cfg-path not set in --local")
			}
			return server.ResetPasswdToLevelDB(
				localCfgPath,
				cctx.String("user"),
				cctx.String("new-passwd"),
			)
		} else {
			ac := client.NewAuthClient(_authApiFlag,
				cctx.String("user"),
				cctx.String("passwd"),
			)
			passwd, err := ac.ResetPasswd(ctx,
				user,
				cctx.String("new-passwd"),
			)
			if err != nil {
				return errors.As(err)
			}
			fmt.Printf("password: %s\n", string(passwd))
			return nil
		}
	},
}
