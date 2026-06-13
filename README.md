# Description
File system implement go io.File with http, auth read/write for userspace

# CMD

Install binary
```
go get github.com/gwaysys/bcstorage
```

Run daemon
```
bcstorage daemon # listen with default value
```

Generate read_write token for space directory
```
bctorage grant public # example user+"/"+public
```

Generate a long time read token for space 'public'
```
bctorage grant --kind=a --exp=0 public # example user+"/"+public
```

Upload to bcstorage
```
bctorage upload --token=$write_token [local path] [remote path] # using the default --user --passwd
```

Download from bcstorage 
```
bctorage download --token=$write_token [local path] [remote path] # using the default --user --passwd
```

# Go client SDK


Generate file token
```
package main

import (
    "fmt"

	"github.com/gwaysys/bcstorage/module/client"
)

func main() {
	ctx := cctx.Context
    authAdrr := "https://127.0.0.1:1330"
    user := "admin" // system username
    passwd := ""    // password for username
	authFile := remotePath
	ac := client.NewAuthClient(
        authAddr,
		cctx.String("user"),
		cctx.String("passwd"),
	)
	newToken, err := ac.NewFileToken(ctx, authFile, cctx.String("kind"), cctx.Int("exp"))
	if err != nil {
		panic(err)
	}
	fmt.Printf("grant:%s, token:%s\n", cctx.String("kind"), newToken)
}
```
