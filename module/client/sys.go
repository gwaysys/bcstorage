package client

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gwaylib/errors"
	"github.com/gwaysys/bcstorage/module/utils"
)

var (
	ErrAuthFailed = errors.New("auth failed")
)

type AuthClient struct {
	Addr   string
	User   string
	Passwd string
}

func NewAuthClient(addr, user, passwd string) *AuthClient {
	// TODO: check the host format
	return &AuthClient{Addr: addr, User: user, Passwd: passwd}
}

// make a ping checking to server
func (auth *AuthClient) Check(ctx context.Context) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", auth.Addr+"/sys/check", nil)
	if err != nil {
		return nil, errors.As(err)
	}
	req.SetBasicAuth(auth.User, auth.Passwd)
	resp, err := utils.HttpsClient.Do(req)
	if err != nil {
		return nil, errors.As(err)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.As(err)
	}
	if resp.StatusCode != 200 {
		return nil, errors.Parse(string(respBody)).As(resp.StatusCode)
	}
	return respBody, nil
}

// add a new user to server, need admin auth
func (auth *AuthClient) AddUser(ctx context.Context, user, passwd string) ([]byte, error) {
	params := make(url.Values)
	params.Add("user", user)
	params.Add("passwd", passwd)
	req, err := http.NewRequestWithContext(ctx, "POST", auth.Addr+"/sys/auth/add", strings.NewReader(params.Encode()))
	if err != nil {
		return nil, errors.As(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(auth.User, auth.Passwd)

	resp, err := utils.HttpsClient.Do(req)
	if err != nil {
		return nil, errors.As(err)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.As(err)
	}
	switch resp.StatusCode {
	case 200:
		return respBody, nil
	case 401:
		// auth failed
		return nil, ErrAuthFailed.As(resp.StatusCode, string(respBody))
	}
	return nil, errors.Parse(string(respBody)).As(resp.StatusCode)
}

// reset special user password, need admin auth
func (auth *AuthClient) ResetPasswd(ctx context.Context, user, passwd string) ([]byte, error) {
	params := make(url.Values)
	params.Add("user", user)
	params.Add("passwd", passwd)
	req, err := http.NewRequestWithContext(ctx, "POST", auth.Addr+"/sys/auth/reset", strings.NewReader(params.Encode()))
	if err != nil {
		return nil, errors.As(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(auth.User, auth.Passwd)

	resp, err := utils.HttpsClient.Do(req)
	if err != nil {
		return nil, errors.As(err)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.As(err)
	}
	switch resp.StatusCode {
	case 200:
		return respBody, nil
	case 401:
		// auth failed
		return nil, ErrAuthFailed.As(resp.StatusCode, string(respBody))
	}
	return nil, errors.Parse(string(respBody)).As(resp.StatusCode)
}

// change passwd of current user
func (auth *AuthClient) ChangePasswd(ctx context.Context, newPasswd string) ([]byte, error) {
	params := make(url.Values)
	params.Add("passwd", newPasswd)
	req, err := http.NewRequestWithContext(ctx, "POST", auth.Addr+"/sys/auth/change", strings.NewReader(params.Encode()))
	if err != nil {
		return nil, errors.As(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(auth.User, auth.Passwd)

	resp, err := utils.HttpsClient.Do(req)
	if err != nil {
		return nil, errors.As(err)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.As(err)
	}
	switch resp.StatusCode {
	case 200:
		return respBody, nil
	case 401:
		// auth failed
		return nil, ErrAuthFailed.As(resp.StatusCode, string(respBody))
	}
	return nil, errors.Parse(string(respBody)).As(resp.StatusCode)
}

// get a temp auth token of file for operating
// the grant == "a", the token WILL NOT BE expiress, and read only, so it can be public(NEED CONFIRM).
// exp -- expire after 'exp' minutes, 0 default to 60 minute
// return the auth token
func (auth *AuthClient) NewFileToken(ctx context.Context, path, grant string, exp int) ([]byte, error) {
	params := url.Values{}
	params.Add("path", path)
	params.Add("grant", grant)
	params.Add("exp", strconv.Itoa(exp))
	req, err := http.NewRequestWithContext(ctx, "POST", auth.Addr+"/sys/file/token?"+params.Encode(), nil)
	if err != nil {
		return nil, errors.As(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(auth.User, auth.Passwd)

	resp, err := utils.HttpsClient.Do(req)
	if err != nil {
		return nil, errors.As(err)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.As(err)
	}
	if resp.StatusCode != 200 {
		return nil, errors.Parse(string(respBody)).As(resp.StatusCode)
	}
	return respBody, nil
}

// release the file auth token manually
func (auth *AuthClient) ReleaseFileToken(ctx context.Context, token string) ([]byte, error) {
	params := url.Values{}
	params.Add("tk", token)
	req, err := http.NewRequestWithContext(ctx, "DELETE", auth.Addr+"/sys/file/token?"+params.Encode(), nil)
	if err != nil {
		return nil, errors.As(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(auth.User, auth.Passwd)

	resp, err := utils.HttpsClient.Do(req)
	if err != nil {
		return nil, errors.As(err)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.As(err)
	}
	if resp.StatusCode != 200 {
		return nil, errors.Parse(string(respBody)).As(resp.StatusCode)
	}
	return respBody, nil
}
