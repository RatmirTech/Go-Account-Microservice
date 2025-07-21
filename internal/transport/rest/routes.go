package rest

import (
	"AccountService/internal/app/account"
	"AccountService/internal/transport/rest/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRouter(accountHandler *account.Handler) *gin.Engine {
	r := gin.Default()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	api := r.Group("/api")
	{
		api.POST("/register", accountHandler.Register)
		api.POST("/login", accountHandler.Login)
		api.POST("/refresh", accountHandler.Refresh)
	}

	auth := r.Group("/api").Use(middleware.AuthMiddleware(accountHandler.Service.Tokens))
	{
		auth.GET("/me", accountHandler.Me)
		auth.POST("/logout", accountHandler.Logout)
	}

	return r
}
