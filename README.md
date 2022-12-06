# gapi

go http api service and client for protobuf

gapi是一个通用的go-http使用protobuf的解决方案，可以兼容[go-gin](https://github.com/gin-gonic/gin)和其他框架，也支持自定义的客户端请求

具体使用，可以参考[gapi-demo](https://github.com/xnzone/gapi-demo)

服务端只需要实现router，就可以无缝对接。如果要使用client，则实现client就可以无缝对接了

同时可以区分http的请求参数，header,uri,query和body