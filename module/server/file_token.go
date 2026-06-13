package server

import (
	"bytes"
	"encoding/json/jsontext"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/gwaylib/errors"
)

const (
	FILE_TOKEN_GRANT_ALL = "a"
)

type FileToken struct {
	_space    string // origin user space
	_file     string // origin file
	grant     string
	expiredAt time.Time

	absSpace string
	absFile  string
}

func NewFileToken(userSpace, filePath, grant string, exp time.Time) (*FileToken, error) {
	absSpacePath, err := filepath.Abs(filepath.Join(_dataPath, userSpace))
	if err != nil {
		return nil, errors.As(err)
	}
	absFilePath, err := filepath.Abs(filepath.Join(_dataPath, userSpace, filePath))
	if err != nil {
		return nil, errors.As(err)
	}
	if !strings.HasPrefix(absFilePath, absSpacePath) {
		return nil, errors.New("file is out of space")
	}
	return &FileToken{
		_space:    userSpace,
		absSpace:  absSpacePath,
		absFile:   absFilePath,
		grant:     grant,
		expiredAt: exp,
	}, nil
}

func (f *FileToken) ReadOnly() bool {
	switch f.grant {
	case FILE_TOKEN_GRANT_ALL:
		return true
	}
	return false
}

// return the authed abs path or empty for error
func (f *FileToken) AuthedPath() string {
	return f.absFile
}

// return the abs path of the file
func (f *FileToken) InAuthPath(file string) (string, bool) {
	absPath, err := filepath.Abs(filepath.Join(f.absSpace, file))
	if err != nil {
		return "", false
	}

	if !strings.HasPrefix(absPath, f.AuthedPath()) {
		return "", false
	}
	return absPath, true
}

func (f *FileToken) UnmarshalJSON(input []byte) error {
	in := bytes.NewReader(input)
	dec := jsontext.NewDecoder(in)
	var _space, _file, grant string
	var expire time.Time
	for {
		// Read a token from the input.
		tok, err := dec.ReadToken()
		if err != nil {
			if err == io.EOF {
				break
			}
			return errors.As(err)
		}
		if tok.Kind() != '"' {
			continue
		}
		switch tok.String() {
		case "_space":
			val, err := dec.ReadToken()
			if err != nil {
				return errors.As(err)
			}
			_space = val.String()
		case "_file":
			val, err := dec.ReadToken()
			if err != nil {
				return errors.As(err)
			}
			_file = val.String()
		case "grant":
			val, err := dec.ReadToken()
			if err != nil {
				return errors.As(err)
			}
			grant = val.String()
		case "exp":
			val, err := dec.ReadToken()
			if err != nil {
				return errors.As(err)
			}
			unixSec := val.Int()
			expire = time.Unix(unixSec, 0)
		}
	}
	// fix abs when abs path changed
	nf, err := NewFileToken(_space, _file, grant, expire)
	if err != nil {
		return errors.As(err)
	}
	*f = *nf
	return nil
}

// json serial
func (f *FileToken) MarshalJSON() ([]byte, error) {
	out := &bytes.Buffer{}
	enc := jsontext.NewEncoder(out, jsontext.Multiline(true)) // expand for readability
	if err := enc.WriteToken(jsontext.BeginObject); err != nil {
		return nil, errors.As(err)
	}
	if err := enc.WriteToken(jsontext.String("_space")); err != nil {
		return nil, errors.As(err)
	}
	if err := enc.WriteToken(jsontext.String(f._space)); err != nil {
		return nil, errors.As(err)
	}
	if err := enc.WriteToken(jsontext.String("_file")); err != nil {
		return nil, errors.As(err)
	}
	if err := enc.WriteToken(jsontext.String(f._file)); err != nil {
		return nil, errors.As(err)
	}
	if err := enc.WriteToken(jsontext.String("exp")); err != nil {
		return nil, errors.As(err)
	}
	if err := enc.WriteToken(jsontext.Int(f.expiredAt.Unix())); err != nil {
		return nil, errors.As(err)
	}
	if err := enc.WriteToken(jsontext.EndObject); err != nil {
		return nil, errors.As(err)
	}
	return out.Bytes(), nil
}
