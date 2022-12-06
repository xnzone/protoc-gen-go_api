package client

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/xnzone/gapi/querystring"
)

type ParseFunc func(ctx context.Context, req interface{}) (header map[string][]string, uri map[string][]string, query map[string][]string, err error)

type CallOptions interface {
	Options() CallOptions
}

type CallOption func(o CallOptions)

type Client interface {
	Do(ctx context.Context, req *http.Request, opts ...CallOption) (*http.Response, error)
}

func Call(ctx context.Context, c Client, req *http.Request, res interface{}, opts ...CallOption) error {
	resp, err := c.Do(ctx, req, opts...)
	if err != nil {
		return err
	}
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}
	var bs []byte
	_, err = resp.Body.Read(bs)
	if err != nil {
		return err
	}
	err = jsoniter.Unmarshal(bs, &res)
	if err != nil {
		return err
	}
	return nil
}

func Parse(_ context.Context, in interface{}) (header map[string][]string, uri map[string][]string, query map[string][]string, err error) {
	if header, err = querystring.Values(in, "header"); err != nil {
		return
	}
	if uri, err = querystring.Values(in, "uri"); err != nil {
		return
	}
	if query, err = querystring.Values(in, "form"); err != nil {
		return
	}
	return
}

func BuildRequest(ctx context.Context, httpMethod string, addr string, relativePath string, in interface{}, body interface{}) (*http.Request, error) {
	header, uri, query, err := Parse(ctx, in)
	if err != nil {
		return nil, err
	}
	for k, v := range uri {
		if len(v) != 1 {
			continue
		}
		relativePath = strings.ReplaceAll(relativePath, fmt.Sprintf(":%s", k), v[0])
		relativePath = strings.ReplaceAll(relativePath, fmt.Sprintf("*%s", k), v[0])
	}
	if len(query) > 0 {
		q := url.Values(query).Encode()
		relativePath = fmt.Sprintf("%s?%s", relativePath, q)
	}
	relativePath = fmt.Sprintf("%s%s", addr, relativePath)
	var bs []byte
	if body != nil {
		bs, _ = jsoniter.Marshal(body)
	}
	req, err := http.NewRequestWithContext(ctx, httpMethod, relativePath, bytes.NewReader(bs))
	if err != nil {
		return nil, err
	}
	for k, v := range header {
		for _, vi := range v {
			req.Header.Add(k, vi)
		}
	}
	return req, nil
}
