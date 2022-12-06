package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/xnzone/gapi/client"
	"github.com/xnzone/gapi/server"
)

type Request struct {
	Body *RequestBody
}

type RequestBody struct {
}

type Respone struct {
}

type Service interface {
	Hello(ctx context.Context, req *Request, res *Respone) error
}

func RegisterServer(router server.Router, srv Service) {
	router.Resolve(http.MethodGet, "/hello", hello(router, srv))
}

func hello(router server.Router, srv Service) func(ctx context.Context) {
	return func(ctx context.Context) {
		in := new(Request)
		if err := router.Bind(ctx, in); err != nil {
			router.Result(ctx, http.StatusInternalServerError, err.Error())
			return
		}
		resp := new(Respone)
		if err := srv.Hello(ctx, in, resp); err != nil {
			router.Result(ctx, http.StatusInternalServerError, err.Error())
			return
		}
		router.Result(ctx, http.StatusOK, resp)
	}
}

type Client interface {
	Hello(ctx context.Context, req *Request) (*Respone, error)
}

type cli struct {
	addr   string
	c      client.Client
	parser client.ParseFunc
}

func (c *cli) Hello(ctx context.Context, req *Request, opts ...client.CallOption) (*Respone, error) {
	request, err := c.buildRequest(ctx, http.MethodPost, "/hello", req, req.Body)
	res := new(Respone)
	if err = client.Call(ctx, c.c, request, res, opts...); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *cli) buildRequest(ctx context.Context, httpMethod string, relativePath string, in interface{}, body interface{}) (*http.Request, error) {
	relativePath = fmt.Sprintf("%s%s", c.addr, relativePath)
	header, uri, query, err := c.parser(ctx, in)
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
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, relativePath, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range header {
		for _, vi := range v {
			request.Header.Add(k, vi)
		}
	}
	return request, nil
}
