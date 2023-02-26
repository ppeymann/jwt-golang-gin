package middleware

import (
	"fmt"
	"jwt/helpers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Authenticate() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		clientToken := ctx.Request.Header.Get("token")
		if clientToken == ""{
			ctx.JSON(http.StatusInternalServerError,gin.H{"err!":fmt.Sprintf("Err")})
			ctx.Abort()
			return
		}
		claims, err := helpers.ValidateToken(clientToken)
		if err != "" {
			ctx.JSON(http.StatusInternalServerError,gin.H{"err":err})
			ctx.Abort()
			return
		}
		ctx.Set("email",claims.Email)
		ctx.Set("first_name",claims.First_name)
		ctx.Set("last_name",claims.Last_name)
		ctx.Set("uid",claims.Uid)
		ctx.Set("user_type",claims.User_type)
		ctx.Next()
	}
}