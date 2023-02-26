package routes

import (
	controllers "jwt/controllers"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(router *gin.Engine){
	router.POST("user/signup",controllers.Signup())
	router.POST("user/login",controllers.Login())
}