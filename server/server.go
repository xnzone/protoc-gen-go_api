package server

import (
	"context"
	"net/http"
)

type HandlerFunc func(ctx context.Context)

type Server interface {
	Resolve(httpMethod string, relativePath string, handlers ...HandlerFunc)
	Bind(ctx context.Context, req interface{}) error
	Result(ctx context.Context, httpStatus int, resp interface{})
}

type CommonHandle func(ctx context.Context, in interface{}, out interface{}) error

func Handle(srv Server, fn CommonHandle, in interface{}, out interface{}) func(ctx context.Context) {
	return func(ctx context.Context) {
		if err := srv.Bind(ctx, in); err != nil {
			srv.Result(ctx, http.StatusInternalServerError, err.Error())
			return
		}
		if err := fn(ctx, in, out); err != nil {
			srv.Result(ctx, http.StatusInternalServerError, err.Error())
			return
		}
		srv.Result(ctx, http.StatusOK, out)
	}
}
