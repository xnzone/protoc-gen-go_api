package server

import "context"

type HandlerFunc func(ctx context.Context)

type Router interface {
	Resolve(httpMethod string, relativePath string, handlers ...HandlerFunc)
	Bind(ctx context.Context, req interface{}) error
	Result(ctx context.Context, httpStatus int, resp interface{})
}
