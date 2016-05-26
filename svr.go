package niuhe

import (
	"github.com/gin-gonic/gin"
	"os"
)

type Server struct {
	modules []*Module
	engine *gin.Engine
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
	svr.GetGinEngine().Run(addr)
}

func (svr *Server) GetGinEngine() *gin.Engine {
	if svr.engine == nil {
		svr.engine = gin.New()
		svr.engine.Use(gin.LoggerWithWriter(os.Stderr), gin.Recovery())
		for _, mod := range svr.modules {
			group := svr.engine.Group(mod.urlPrefix)
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
