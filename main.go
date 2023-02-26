package main

import (
	"jwt/routes"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	port := os.Getenv("PORT")

	if port==""{
		port = "8000"
	}
	router.Use(gin.Logger())
	
	// routes.AuthRoutes(router)
	// routes.UserRoutes(router)
	
	//* API
	router.GET("/api-1",func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK,gin.H{"success":"access to API-1"})
	})
	router.GET("/api-2",func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK,gin.H{"success":"access to API-2"})
	})
	routes.AuthRoutes(router)
	routes.UserRoutes(router)


	router.Run( ":" + port)
}