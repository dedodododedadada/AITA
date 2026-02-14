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
	router.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
	
	v1 := router.Group("/api/v1")
	{
		v1.POST("/signup", userHandler.SignUp)
		v1.POST("/login", userHandler.Login)
		v1.GET("/tweets/:id", tweetHandler.Get)
		protected := v1.Group("/")
		protected.Use(AuthMiddleware(sessionService))
		{
			protected.GET("/me", userHandler.GetMe)
			protected.POST("/logout", userHandler.Logout)
			tweets := protected.Group("/tweets")
			{
				tweets.POST("", tweetHandler.Create)
				tweets.PATCH("/:id", tweetHandler.Update)  
                tweets.DELETE("/:id", tweetHandler.Delete)
			}
		} 
	}
	return router
}
