package client

import (
	"context"
	"net/http"

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

func Parse(ctx context.Context, in interface{}) (header map[string][]string, uri map[string][]string, query map[string][]string, err error) {
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
