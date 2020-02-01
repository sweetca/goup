// Copyright (c) 2020 Hervé Gouchet. All rights reserved.
// Use of this source code is governed by the MIT License
// that can be found in the LICENSE file.

package goget

import (
	"context"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/rvflash/goup"
	"github.com/rvflash/goup/internal/semver"
	"github.com/rvflash/goup/internal/vcs"
	"github.com/rvflash/goup/internal/vcs/git"
)

// Name is the name of this VCS.
const Name = "go-get"

// System
type System struct {
	client vcs.HTTPClient
	git    *git.System
}

// New
func New(client vcs.HTTPClient, git *git.System) *System {
	return &System{
		client: client,
		git:    git,
	}
}

// CanFetch implements the vcs.System interface.
func (s *System) CanFetch(path string) bool {
	return strings.Contains(path, "golang.org")
}

// FetchPath implements the vcs.System interface.
func (s *System) FetchPath(ctx context.Context, path string) (semver.Tags, error) {
	system, remote, err := s.vcsByPath(ctx, path)
	if err != nil {
		return nil, err
	}
	return s.fetchURL(ctx, system, remote)
}

// FetchURL implements the vcs.System interface.
func (s *System) FetchURL(ctx context.Context, url string) (semver.Tags, error) {
	system, remote, err := s.vcsByURL(ctx, url)
	if err != nil {
		return nil, err
	}
	return s.fetchURL(ctx, system, remote)
}

func (s *System) fetchURL(ctx context.Context, system, url string) (semver.Tags, error) {
	if s.git == nil {
		return nil, goup.ErrSystem
	}
	switch system {
	case git.Name:
		return s.git.FetchURL(ctx, url)
	default:
		return nil, goup.ErrSystem
	}
}

func (s *System) vcsByURL(ctx context.Context, url string) (vcs, remote string, err error) {
	if ctx == nil || s.client == nil {
		return "", "", goup.ErrSystem
	}
	if url == "" {
		return "", "", goup.ErrRepository
	}
	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return
	}
	var resp *http.Response
	resp, err = s.client.Do(req)
	if err != nil {
		return
	}
	defer func() { _ = resp.Body.Close() }()
	return parseMetaGoImport(resp.Body)
}

func (s *System) vcsByPath(ctx context.Context, path string) (vcs, remote string, err error) {
	if path == "" {
		return "", "", goup.ErrRepository
	}
	for _, protocol := range []string{"https://", "http://"} {
		vcs, remote, err = s.vcsByURL(ctx, protocol+path)
		if err == nil {
			break
		}
	}
	return
}

const (
	body    = "body"
	head    = "head"
	meta    = "meta"
	name    = "name"
	attr    = "go-import"
	content = "content"

	utf8Charset  = "utf-8"
	asciiCharset = "ascii"
)

func parseMetaGoImport(r io.Reader) (vcs, url string, err error) {
	d := xml.NewDecoder(r)
	d.CharsetReader = charsetReader
	d.Strict = false
	for {
		t, err := d.RawToken()
		if err != nil {
			break
		}
		if e, ok := t.(xml.StartElement); ok && strings.EqualFold(e.Name.Local, body) {
			break
		}
		if e, ok := t.(xml.EndElement); ok && strings.EqualFold(e.Name.Local, head) {
			break
		}
		e, ok := t.(xml.StartElement)
		if !ok || !strings.EqualFold(e.Name.Local, meta) {
			continue
		}
		if attrValue(e.Attr, name) != attr {
			continue
		}
		if f := strings.Fields(attrValue(e.Attr, content)); len(f) == 3 {
			vcs = f[1]
			url = f[2]
			break
		}
	}
	return
}

func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	switch strings.ToLower(charset) {
	case utf8Charset, asciiCharset:
		return input, nil
	default:
		return nil, goup.Errorf(Name, errors.New("charset: "+charset))
	}
}

func attrValue(attrs []xml.Attr, name string) string {
	for _, a := range attrs {
		if strings.EqualFold(a.Name.Local, name) {
			return a.Value
		}
	}
	return ""
}
