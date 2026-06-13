package client

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/gwaylib/errors"
	"github.com/gwaylib/log"
	"github.com/gwaysys/bcstorage/module/utils"
)

var (
	ErrHashNotMatch = errors.New("hash not match")
	_errNotExist    error
)

func init() {
	// init the File not exist cause by can't get the error in os package
	for {
		if _errNotExist != nil {
			break
		}
		_, err := os.Stat(uuid.New().String())
		if err == nil {
			continue
		}
		pErr, ok := err.(*os.PathError)
		if !ok {
			panic(err)
		}
		_errNotExist = pErr.Err
	}
}

type FileInfo struct {
	name     string
	size     int64
	fileMode os.FileMode
	modTime  time.Time
	isDir    bool
}

func (fInfo *FileInfo) Name() string {
	// base name of the file
	return fInfo.name
}
func (fInfo *FileInfo) Size() int64 {
	// length in bytes for regular files; system-dependent for others
	return fInfo.size
}
func (fInfo *FileInfo) Mode() os.FileMode {
	// file mode bits
	return fInfo.fileMode
}
func (fInfo *FileInfo) ModTime() time.Time {
	// modification time
	return fInfo.modTime
}
func (fInfo *FileInfo) IsDir() bool {
	// abbreviation for Mode().IsDir()
	return fInfo.isDir
}
func (fInfo *FileInfo) Sys() interface{} {
	// underlying data source (can return nil)
	return nil
}

type HttpClient struct {
	Addr      string
	AuthToken string
}

// AuthSpace is must need.
// AuthPath is a scope control.
func NewHttpClient(addr, authToken string) *HttpClient {
	return &HttpClient{Addr: addr, AuthToken: authToken}
}

func (f *HttpClient) Capacity(ctx context.Context) (*syscall.Statfs_t, error) {
	params := url.Values{}
	params.Add("tk", f.AuthToken)
	req, err := http.NewRequestWithContext(ctx, "GET", f.Addr+"/file/capacity?"+params.Encode(), nil)
	if err != nil {
		return nil, errors.As(err)
	}
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
	st := &syscall.Statfs_t{}
	if err := json.Unmarshal(respBody, st); err != nil {
		return nil, errors.As(err)
	}
	return st, nil

}

// the path need be a part of AuthPath
func (f *HttpClient) Move(ctx context.Context, oldRemotePath, newRemotePath string) error {
	params := url.Values{}
	params.Add("tk", f.AuthToken)
	params.Add("ta", oldRemotePath)
	params.Add("new", newRemotePath)
	req, err := http.NewRequestWithContext(ctx, "POST", f.Addr+"/file/move?"+params.Encode(), nil)
	if err != nil {
		return errors.As(err, oldRemotePath, newRemotePath)
	}
	resp, err := utils.HttpsClient.Do(req)
	if err != nil {
		return errors.As(err, oldRemotePath, newRemotePath)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.As(err)
	}
	if resp.StatusCode != 200 {
		return errors.Parse(string(respBody)).As(resp.StatusCode)
	}
	return nil
}

// the file need be a part of AuthPath
func (f *HttpClient) Delete(ctx context.Context, file string) error {
	params := url.Values{}
	params.Add("tk", f.AuthToken)
	params.Add("ta", file)
	req, err := http.NewRequestWithContext(ctx, "POST", f.Addr+"/file/delete?"+params.Encode(), nil)
	if err != nil {
		return errors.As(err)
	}
	resp, err := utils.HttpsClient.Do(req)
	if err != nil {
		return errors.As(err)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.As(err)
	}
	if resp.StatusCode != 200 {
		return errors.Parse(string(respBody)).As(resp.StatusCode)
	}
	return nil
}

// the remoteFile need a sub path of AuthPath
func (f *HttpClient) Truncate(ctx context.Context, remoteFile string, size int64) error {
	params := url.Values{}
	params.Add("tk", f.AuthToken)
	params.Add("ta", remoteFile)
	params.Add("size", strconv.FormatInt(size, 10))
	req, err := http.NewRequestWithContext(ctx, "GET", f.Addr+"/file/truncate?"+params.Encode(), nil)
	if err != nil {
		return errors.As(err)
	}
	resp, err := utils.HttpsClient.Do(req)
	if err != nil {
		return errors.As(err)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.As(err)
	}
	if resp.StatusCode != 200 {
		return errors.Parse(string(respBody)).As(resp.StatusCode)
	}
	return nil
}

// the remotePath need be a part of sub AuthPath
func (f *HttpClient) FileStat(ctx context.Context, remotePath string) (os.FileInfo, error) {
	params := url.Values{}
	params.Add("tk", f.AuthToken)
	params.Add("ta", remotePath)
	req, err := http.NewRequest("GET", f.Addr+"/file/stat?"+params.Encode(), nil)
	if err != nil {
		return nil, errors.As(err)
	}
	resp, err := utils.HttpsClient.Do(req)
	if err != nil {
		return nil, errors.As(err)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.As(err)
	}
	// file not found
	switch resp.StatusCode {
	case 404:
		return nil, &os.PathError{Op: "FileStat", Path: remotePath, Err: _errNotExist}
	case 200:
		stat := &utils.ServerFileStat{}
		if err := json.Unmarshal(respBody, &stat); err != nil {
			return nil, errors.As(err)
		}
		stat.FileName = remotePath
		return stat, nil

	default:
		return nil, errors.Parse(resp.Status).As(resp.StatusCode, string(respBody))
	}
}

// TODO: erasure coding
// the remotePath need be a sub part of AuthPath
func (f *HttpClient) write(ctx context.Context, localPath, remotePath string, append bool) (int64, error) {
	pos := int64(0)
	if append {
		// Get the file information
		info, err := f.FileStat(ctx, remotePath)
		if err != nil {
			if !os.IsNotExist(err) {
				return 0, errors.As(err)
			}
			info = &FileInfo{name: remotePath}
		}
		if info.Size() > 0 {
			pos = info.Size() - 1
		}
	} else {
		if err := f.Truncate(ctx, remotePath, 0); err != nil {
			return 0, errors.As(err)
		}
	}

	localFile, err := os.Open(localPath)
	if err != nil {
		return 0, errors.As(err)
	}
	defer localFile.Close()

	if _, err := localFile.Seek(pos, 0); err != nil {
		return 0, errors.As(err)
	}
	localStat, err := localFile.Stat()
	if err != nil {
		return 0, errors.As(err)
	}

	// get remote io
	params := url.Values{}
	params.Add("tk", f.AuthToken)
	params.Add("ta", remotePath)
	params.Add("pos", strconv.FormatInt(pos, 10))
	params.Add("checksum", "sha1")
	req, err := http.NewRequestWithContext(ctx, "POST", f.Addr+"/file/write?"+params.Encode(), localFile)
	if err != nil {
		return 0, errors.As(err)
	}
	resp, err := utils.HttpsClient.Do(req)
	if err != nil {
		return 0, errors.As(err)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, errors.As(err)
	}
	if resp.StatusCode != 200 {
		return 0, errors.Parse(string(respBody)).As(resp.StatusCode)
	}

	// checksum the upload file.
	localFile.Close()
	localFile, err = os.Open(localPath)
	if err != nil {
		return 0, errors.As(err)
	}
	defer localFile.Close()
	localHash := sha1.New()
	if _, err := localFile.Seek(0, 0); err != nil {
		return 0, errors.As(err)
	}
	if _, err := io.Copy(localHash, localFile); err != nil {
		return 0, errors.As(err, localPath)
	}
	localSum := fmt.Sprintf("%x", localHash.Sum(nil))

	if localSum != string(respBody) {
		log.Warnf("upload file not match, retransmit %s:%s,%s", remotePath, localSum, string(respBody))
		if !append {
			return 0, ErrHashNotMatch.As(string(respBody), localSum)
		}

		// try again
		return f.write(ctx, localPath, remotePath, false)
	}

	if append {
		return localStat.Size() - (pos + 1), nil
	}
	return localStat.Size() - pos, nil
}

// the remotePath need be a sub part of AuthPath
func (f *HttpClient) Upload(ctx context.Context, localPath, remotePath string) error {
	fStat, err := os.Lstat(localPath)
	if err != nil {
		return errors.As(err, localPath)
	}
	if !fStat.IsDir() {
		if _, err := f.write(ctx, localPath, remotePath, true); err != nil {
			return errors.As(err)
		}
		return nil
	}

	dirs, err := ioutil.ReadDir(localPath)
	if err != nil {
		return errors.As(err)
	}
	for _, fs := range dirs {
		newLocalPath := filepath.Join(localPath, fs.Name())
		newRemotePath := filepath.Join(remotePath, fs.Name())
		if fs.IsDir() {
			if err := f.Upload(ctx, newLocalPath, newRemotePath); err != nil {
				return errors.As(err)
			}
			continue
		}
		if _, err := f.write(ctx, newLocalPath, newRemotePath, true); err != nil {
			return errors.As(err)
		}
	}
	return nil
}

func (f *HttpClient) List(ctx context.Context, remotePath string) ([]utils.ServerFileStat, error) {
	params := url.Values{}
	params.Add("tk", f.AuthToken)
	params.Add("ta", remotePath)
	req, err := http.NewRequestWithContext(ctx, "GET", f.Addr+"/file/list?"+params.Encode(), nil)
	if err != nil {
		return nil, errors.As(err, remotePath)
	}
	resp, err := utils.HttpsClient.Do(req)
	if err != nil {
		return nil, errors.As(err, remotePath)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.As(err, remotePath)
	}
	switch resp.StatusCode {
	case 200:
		list := []utils.ServerFileStat{}
		if err := json.Unmarshal(respBody, &list); err != nil {
			return nil, errors.As(err, remotePath)
		}
		return list, nil
	case 404:
		return nil, &os.PathError{Op: "List", Path: remotePath, Err: _errNotExist}
	}
	return nil, errors.Parse(string(respBody)).As(resp.StatusCode)
}

// TODO: erasure coding
func (f *HttpClient) read(ctx context.Context, localPath, remotePath string) (int64, error) {
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return 0, errors.As(err)
	}
	toFile, err := os.OpenFile(localPath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return 0, errors.As(err, localPath)
	}
	defer toFile.Close()
	toStat, err := toFile.Stat()
	if err != nil {
		return 0, errors.As(err, localPath)
	}
	pos := int64(0)
	if toStat.Size() > 0 {
		pos = toStat.Size() - 1
	}

	params := url.Values{}
	params.Add("tk", f.AuthToken)
	params.Add("ta", remotePath)
	req, err := http.NewRequestWithContext(ctx, "GET", f.Addr+"/file/read?"+params.Encode(), nil)
	if err != nil {
		return 0, errors.As(err)
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-", pos))
	resp, err := utils.HttpsClient.Do(req)
	if err != nil {
		return 0, errors.As(err)
	}
	defer resp.Body.Close()

	// file not found
	switch resp.StatusCode {
	case 200, 206:
		// continue
	case 404:
		return 0, &os.PathError{Op: "read", Path: remotePath, Err: _errNotExist}
	case 416:
		return 0, io.EOF
	default:
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return 0, errors.As(err)
		}
		return 0, errors.New(resp.Status).As(resp.StatusCode, pos, string(respBody))
	}

	if _, err := toFile.Seek(pos, 0); err != nil {
		return 0, errors.As(err)
	}
	n, err := io.Copy(toFile, resp.Body)
	if err != nil {
		return 0, errors.As(err)
	}
	if pos > 0 {
		return n - 1, nil
	}
	return n, nil
}

func (f *HttpClient) Download(ctx context.Context, localPath, remotePath string) error {
	sFiles, err := f.List(ctx, remotePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return errors.As(err)
		}
		// keep the origin format for file not exist.
		return err
	}
	for _, sf := range sFiles {
		// special protocal for remotePath is file.
		if sf.FileName == "." {
			if _, err := f.read(ctx, localPath, remotePath); err != nil {
				return errors.As(err, localPath, remotePath)
			}
			continue
		}

		newLocalPath := filepath.Join(localPath, sf.FileName)
		newRemotePath := filepath.Join(remotePath, sf.FileName)
		if sf.IsDirFile {
			if err := f.Download(ctx, newLocalPath, newRemotePath); err != nil {
				return errors.As(err, newLocalPath, newRemotePath)
			}
			continue
		}
		if _, err := f.read(ctx, newLocalPath, newRemotePath); err != nil {
			return errors.As(err, newLocalPath, newRemotePath)
		}
	}
	return nil
}

// TODO: redesign read and write.
//
// implement os.File interface
type HttpFile struct {
	ctx    context.Context
	client *HttpClient

	filePath   string
	lock       sync.Mutex
	seekOffset int64
}

func OpenHttpFile(ctx context.Context, host, remotePath, authToken string) *HttpFile {
	return &HttpFile{
		ctx:      ctx,
		client:   NewHttpClient(host, authToken),
		filePath: remotePath,
	}
}

func (f *HttpFile) Name() string {
	return f.filePath
}

func (f *HttpFile) readRemote(b []byte, off int64) (int, error) {
	params := url.Values{}
	params.Add("tk", f.client.AuthToken)
	params.Add("ta", f.filePath)
	req, err := http.NewRequestWithContext(f.ctx, "GET", f.client.Addr+"/file/read?"+params.Encode(), nil)
	if err != nil {
		return 0, errors.As(err)
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", off, off+int64(len(b))))
	log.Debugf("Range:%s", fmt.Sprintf("bytes=%d-%d", off, off+int64(len(b))))
	resp, err := utils.HttpsClient.Do(req)
	if err != nil {
		return 0, errors.As(err)
	}
	defer resp.Body.Close()

	// file not found
	switch resp.StatusCode {
	case 200, 206:
		// continue
	case 404:
		return 0, &os.PathError{Op: "readRemote", Path: f.filePath, Err: _errNotExist}
	case 416:
		return 0, io.EOF
	default:
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return 0, errors.As(err)
		}
		return 0, errors.Parse(string(respBody)).As(resp.StatusCode)
	}
	n, err := resp.Body.Read(b)
	f.seekOffset = off + int64(n)
	return n, err
}
func (f *HttpFile) writeRemote(b []byte, off int64) (int64, error) {
	r := bytes.NewReader(b)
	params := url.Values{}
	params.Add("token", f.client.AuthToken)
	params.Add("file", f.filePath)
	params.Add("pos", strconv.FormatInt(off, 10))
	req, err := http.NewRequestWithContext(f.ctx, "POST", f.client.Addr+"/file/write?"+params.Encode(), r)
	if err != nil {
		return 0, errors.As(err)
	}
	resp, err := utils.HttpsClient.Do(req)
	if err != nil {
		return 0, errors.As(err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200, 206:
		// continue
	case 404:
		return 0, &os.PathError{Op: "writeRemote", Path: f.filePath, Err: _errNotExist}
	default:
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return 0, errors.As(err)
		}
		return 0, errors.Parse(string(respBody)).As(resp.StatusCode)
	}
	f.seekOffset = off + int64(len(b))
	return int64(len(b)), nil
}

func (f *HttpFile) Close() error {
	return nil
}

func (f *HttpFile) Seek(offset int64, whence int) (ret int64, err error) {
	if whence != 0 {
		return 0, errors.New("unsupport whence not zero")
	}

	f.lock.Lock()
	defer f.lock.Unlock()
	f.seekOffset = offset
	return f.seekOffset, nil
}

func (f *HttpFile) read(b []byte) (n int, err error) {
	written, err := f.readRemote(b, f.seekOffset)
	if err != nil {
		return int(written), err
	}
	if written < len(b) {
		return int(written), io.EOF
	}
	return int(written), err
}
func (f *HttpFile) Read(b []byte) (n int, err error) {
	f.lock.Lock()
	defer f.lock.Unlock()

	n, err = f.read(b)
	if err != nil {
		if io.EOF != err {
			return n, err
		}
	}
	return n, nil

}
func (f *HttpFile) ReadAt(b []byte, off int64) (n int, err error) {
	if off < 0 {
		return 0, &os.PathError{Op: "readat", Path: f.filePath, Err: errors.New("negative offset")}
	}

	f.lock.Lock()
	defer f.lock.Unlock()
	f.seekOffset = off

	n, err = f.read(b)
	if n < len(b) {
		return n, err
	}
	return n, nil
}

func (f *HttpFile) Write(b []byte) (n int, err error) {
	f.lock.Lock()
	defer f.lock.Unlock()

	written, err := f.writeRemote(b, f.seekOffset)
	return int(written), err
}

func (f *HttpFile) Stat() (os.FileInfo, error) {
	return f.client.FileStat(f.ctx, f.filePath)
}

// remotePath need be a part of AuthPath
func (f *HttpFile) Truncate(remotePath string, size int64) error {
	return f.client.Truncate(f.ctx, remotePath, size)
}
