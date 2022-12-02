package niuhe

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	PathPrefix         string
	engine             *gin.Engine
	modules            []*Module
	middlewares        []gin.HandlerFunc
	niuheMiddlewares   []HandlerFunc
	staticPaths        []staticPath
	customLogFormatter func(param gin.LogFormatterParams) string
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
		filename := addr[5:]
		go func() {
			time.Sleep(1 * time.Second)
			os.Chmod(filename, 0777)
		}()
		ginEngine.RunUnix(filename)
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

func (svr *Server) SetCustomLogFormatter(formatter func(gin.LogFormatterParams) string) {
	svr.customLogFormatter = formatter
}

func (svr *Server) logFormatter(param gin.LogFormatterParams) string {
	if svr.customLogFormatter != nil {
		return svr.customLogFormatter(param)
	}
	var statusColor, methodColor, resetColor string
	if param.IsOutputColor() {
		statusColor = param.StatusCodeColor()
		methodColor = param.MethodColor()
		resetColor = param.ResetColor()
	}

	if param.Latency > time.Minute {
		param.Latency = param.Latency.Truncate(time.Second)
	}
	return fmt.Sprintf("[GIN] %v |%s %3d %s| %13v | %15s |%s %-7s %s %#v\n%s",
		param.TimeStamp.Format("2006/01/02 - 15:04:05"),
		statusColor, param.StatusCode, resetColor,
		param.Latency,
		param.ClientIP,
		methodColor, param.Method, resetColor,
		param.Path,
		param.ErrorMessage,
	)
}

func (svr *Server) GetGinEngine(loggerConfig ...gin.LoggerConfig) *gin.Engine {
	if svr.engine == nil {
		svr.engine = gin.New()
		for _, sp := range svr.staticPaths {
			svr.engine.Static(sp.relativePath, sp.root)
		}
		loggerConfig := gin.LoggerConfig{
			Formatter: svr.logFormatter,
			Output:    os.Stderr,
		}
		svr.engine.Use(gin.LoggerWithConfig(loggerConfig), gin.Recovery()).
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
