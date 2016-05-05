package niuhe

import (
	"github.com/gin-gonic/gin"
)

type Server struct {
	modules []*Module
}

func NewServer() *Server {
	return &Server{
		modules: make([]*Module, 0),
	}
}

func (svr *Server) RegisterModule(mod *Module) {
	svr.modules = append(svr.modules, mod)
}

func (svr *Server) Serve(addr string) {
	gin := gin.New()
	for _, mod := range svr.modules {
		group := gin.Group(mod.urlPrefix)
		for _, info := range mod.handlers {
			group.GET(info.Path, info.handleFunc)
			group.POST(info.Path, info.handleFunc)
		}
	}
	gin.Run(addr)
}
