package niuhe

import (
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type Server struct {
	PathPrefix       string
	engine           *gin.Engine
	modules          []*Module
	middlewares      []gin.HandlerFunc
	niuheMiddlewares []HandlerFunc
	staticPaths      []staticPath
}

func NewServer() *Server {
	return &Server{
		modules:          make([]*Module, 0),
		middlewares:      make([]gin.HandlerFunc, 0),
		niuheMiddlewares: make([]HandlerFunc, 0),
		staticPaths:      make([]staticPath, 0),
	}
}

func (svr *Server) Use(middlewares ...gin.HandlerFunc) *Server {
	svr.middlewares = append(svr.middlewares, middlewares...)
	return svr
}

func (svr *Server) UseNiuhe(middlewares ...HandlerFunc) *Server {
	svr.niuheMiddlewares = append(svr.niuheMiddlewares, middlewares...)
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
	ginEngine := svr.GetGinEngine()
	if strings.HasPrefix(addr, "unix:") {
		ginEngine.RunUnix(addr[5:])
	} else {
		ginEngine.Run(addr)
	}
}

type staticPath struct {
	relativePath string
	root         string
}

func (svr *Server) Static(relativePath, root string) {
	svr.staticPaths = append(svr.staticPaths, staticPath{relativePath, root})
}

func (svr *Server) GetGinEngine() *gin.Engine {
	if svr.engine == nil {
		svr.engine = gin.New()
		for _, sp := range svr.staticPaths {
			svr.engine.Static(sp.relativePath, sp.root)
		}
		svr.engine.Use(gin.LoggerWithWriter(os.Stderr), gin.Recovery()).
			Use(svr.middlewares...)
		for _, mod := range svr.modules {
			group := svr.engine.Group(svr.PathPrefix + mod.urlPrefix)
			for _, info := range mod.Routers(svr.niuheMiddlewares) {
				if (info.Methods & GET) != 0 {
					group.GET(info.Path, info.HandleFunc)
				}
				if (info.Methods & POST) != 0 {
					group.POST(info.Path, info.HandleFunc)
				}
			}
		}
	}
	return svr.engine
}
