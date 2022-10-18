package moxio

import (
	"context"
	"encoding/json"
	"github.com/lithictech/moxpopuli/internal"
	"github.com/lithictech/moxpopuli/moxjson"
	"github.com/pkg/errors"
	"io"
	"net/url"
	"os"
)

func Save(ctx context.Context, uri, arg string, i interface{}) error {
	saver, err := NewSaver(ctx, uri, arg)
	if err != nil {
		return errors.Wrap(err, "creating object saver")
	}
	if err := saver.Save(ctx, i); err != nil {
		return errors.Wrap(err, "saving object saver")
	}
	return nil
}

type Saver interface {
	Save(context.Context, interface{}) error
}

func NewSaver(_ context.Context, uri, arg string) (Saver, error) {
	if uri == "" || uri == "-" {
		return streamSaver{w: os.Stdout}, nil
	}
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "file" {
		return fileSaver{filePath: internal.FileUriPath(u), jsonPath: arg}, nil
	}
	return nil, errors.Errorf("unknown saver for scheme '%s'", u.Scheme)
}

type fileSaver struct {
	filePath string
	jsonPath string
}

func (s fileSaver) Save(_ context.Context, i interface{}) error {
	var toEncode interface{}
	if s.jsonPath == "" {
		toEncode = i
	} else {
		existingF, err := os.Open(s.filePath)
		if os.IsNotExist(err) {
			// Nothing to load.
		} else if err != nil {
			return errors.Wrap(err, "opening file for modification")
		} else {
			defer existingF.Close()
			if err := json.NewDecoder(existingF).Decode(&toEncode); err == io.EOF {
				// File is empty so just skip it
			} else if err != nil {
				return errors.Wrap(err, "decoding existing file")
			} else {
				pathT := moxjson.ParsePath(s.jsonPath)
				if err := moxjson.Set(toEncode, i, pathT); err != nil {
					return errors.Wrap(err, "setting field in loaded file")
				}
			}
		}
	}

	f, err := os.Create(s.filePath)
	if err != nil {
		return errors.Wrap(err, "recreating file")
	}
	defer f.Close()
	if err := moxjson.NewPrettyEncoder(f).Encode(toEncode); err != nil {
		return errors.Wrap(err, "writing saver")
	}
	return nil

}

type streamSaver struct {
	w io.Writer
}

func (s streamSaver) Save(_ context.Context, i interface{}) error {
	if err := moxjson.NewPrettyEncoder(s.w).Encode(i); err != nil {
		return errors.Wrap(err, "writing saver")
	}
	return nil
}
