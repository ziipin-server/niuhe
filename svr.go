package niuhe

import (
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type Server struct {
	PathPrefix  string
	engine      *gin.Engine
	modules     []*Module
	middlewares []gin.HandlerFunc
}

func NewServer() *Server {
	return &Server{
		modules:     make([]*Module, 0),
		middlewares: make([]gin.HandlerFunc, 0),
	}
}

func (svr *Server) Use(middlewares ...gin.HandlerFunc) *Server {
	svr.middlewares = append(svr.middlewares, middlewares...)
	return svr
}

func (svr *Server) SetPathPrefix(prefix string) {
	if prefix != "" {
		if !strings.HasSuffix(prefix, "/") {
			prefix = fmt.Sprintf("%s/", prefix)
		}
	}
	svr.PathPrefix = prefix
}

func (svr *Server) RegisterModule(mod *Module) {
	svr.modules = append(svr.modules, mod)
}

func (svr *Server) Serve(addr string) {
	svr.GetGinEngine().Run(addr)
}

func (svr *Server) GetGinEngine() *gin.Engine {
	if svr.engine == nil {
		svr.engine = gin.New()
		svr.engine.Use(gin.LoggerWithWriter(os.Stderr), gin.Recovery()).
			Use(svr.middlewares...)
		for _, mod := range svr.modules {
			group := svr.engine.Group(svr.PathPrefix + mod.urlPrefix)
			for _, info := range mod.handlers {
				if (info.Methods & GET) != 0 {
					group.GET(info.Path, info.handleFunc)
				}
				if (info.Methods & POST) != 0 {
					group.POST(info.Path, info.handleFunc)
				}
			}
		}
	}
	return svr.engine
}
