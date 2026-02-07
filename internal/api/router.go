package api

import (
	"github.com/gin-gonic/gin"
)

func SetupRouter(
	userHandler *UserHandler, 
	tweetHandler *TweetHandler, 
	sessionService AuthSessionService ,
	) *gin.Engine {
	router := gin.Default()
	v1 := router.Group("/api/v1")
	{
		v1.POST("/signup", userHandler.SignUp)
		v1.POST("/login", userHandler.Login)
		protected := v1.Group("/")
		protected.Use(AuthMiddleware(sessionService))
		{
			protected.GET("/me", userHandler.GetMe)
			tweets := protected.Group("/tweets")
			{
				tweets.POST("", tweetHandler.Create)
			}
		} 
	}
	return router
}
