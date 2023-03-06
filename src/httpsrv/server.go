package httpsrv

import (
	"1c-bitrix-mediator/src/httpsrv/routes"
	"fmt"

	"github.com/gin-gonic/gin"
)

type HttpServer struct {
	HttpServer *gin.Engine
}

func InitHttpServer() HttpServer {
	router := gin.Default()
	return HttpServer{router}
}

func (s *HttpServer) Run(host string, port string) {
	s.HttpServer.Run(fmt.Sprintf("%s:%s", host, port))
}

func (s *HttpServer) BuildRecordRoutes() {
	prefix := "/document"
	s.HttpServer.POST(prefix, routes.CreateDocument)
}
