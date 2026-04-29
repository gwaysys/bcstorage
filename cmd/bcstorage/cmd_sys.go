package main

import (
	"fmt"
	"path/filepath"

	"github.com/gwaysys/bcstorage/lib/bcrypt"
	"github.com/gwaysys/bcstorage/module/client"
	"github.com/gwaylib/errors"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"
)

var sysCmd = &cli.Command{
	Name: "sys",
	Subcommands: []*cli.Command{
		statusCmd,
		authNewCmd,
		authResetCmd,
	},
}
var statusCmd = &cli.Command{
	Name:  "status",
	Usage: "status of storage",
	Action: func(cctx *cli.Context) error {
		// TODO: process the download
		ctx := cctx.Context
		ac := client.NewAuthClient(_authApiFlag,
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
var authNewCmd = &cli.Command{
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
var authResetCmd = &cli.Command{
	Name:  "reset-passwd",
	Usage: "reset user passwd",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "local",
			Value: false,
			Usage: "reset password to leveldb directly after stop the daemon",
		},
		&cli.StringFlag{
			Name:  "repo",
			Value: "",
			Usage: "repo in local mode",
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
			repo, err := homedir.Expand(cctx.String("repo"))
			if err != nil {
				return errors.As(err)
			}
			if len(repo) == 0 {
				return errors.New("repo not set in --local")
			}
			if err := InitLevelDB(filepath.Join(repo, _leveldbPath)); err != nil {
				return errors.As(err)
			}
			defer CloseLevelDB()
			fmt.Println("leveldb is load: ", _levelDB != nil)

			userAuth := UserAuth{}
			userKey := fmt.Sprintf(_leveldb_prefix_user, user)
			if err := ScanLevelDB(userKey, &userAuth); err != nil {
				return errors.As(err)
			}
			passwd := genPasswd()
			userAuth.Passwd = bcrypt.BcryptPwd(passwd)
			if err := PutLevelDB(userKey, &userAuth); err != nil {
				return errors.As(err)
			}
			fmt.Printf("password: %s\n", passwd)
			return nil
		} else {
			ac := client.NewAuthClient(_authApiFlag,
				cctx.String("user"),
				cctx.String("passwd"),
			)
			passwd, err := ac.ResetUserPasswd(ctx, user)
			if err != nil {
				return errors.As(err)
			}
			fmt.Printf("password: %s\n", string(passwd))
			return nil
		}
		return nil
	},
}
