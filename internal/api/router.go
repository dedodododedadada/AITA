package api

import(
	"github.com/gin-gonic/gin"
	"aita/internal/db"
)

func SetupRouter(userHandler *UserHandler, sessionStore db.SessionStore) *gin.Engine {
	router := gin.Default()
	v1 := router.Group("/api/v1")
	{
		v1.POST("/signup", userHandler.SignUp)
		v1.POST("/login", userHandler.Login)
		protected := v1.Group("/")
		protected.Use(AuthMiddleware(sessionStore))
		{
			protected.GET("/me", userHandler.GetMe)
		} 
	}
	return router
}
