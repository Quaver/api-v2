package main

import (
	"fmt"
	"github.com/Quaver/api2/handlers"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
)

// Starts the server on a given port
func initializeServer(port int) {
	gin.SetMode(gin.ReleaseMode)
	logrus.Info("Starting gin server in mode: ", gin.Mode())

	engine := gin.New()
	engine.Use(gin.Recovery())
	initializeRoutes(engine)

	logrus.Info(fmt.Sprintf("API is now being served on port :%v", port))
	logrus.Fatal(engine.Run(fmt.Sprintf(":%v", port)))
}

// Initializes all the routes for the server.
func initializeRoutes(engine *gin.Engine) {
	//engine.GET(createV2Route("/clans"))
	//
	//engine.POST(createV2Route("/clan"))
	//engine.GET(createV2Route("/clan/:id"))
	//engine.PATCH(createV2Route("/clan/:id"))
	//engine.DELETE(createV2Route("/clan/:id"))

	engine.NoRoute(func(c *gin.Context) {
		handlers.ReturnError(c, http.StatusNotFound, "Not Found")
	})

	logrus.Info("Initialized router")
}

func createV2Route(route string) string {
	return fmt.Sprintf("/v2%v", route)
}
