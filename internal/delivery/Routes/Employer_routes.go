package routes

import (
	"github.com/gin-gonic/gin"
	handler "skillsync-authservice/internal/delivery/handler"
)

func RegisterEmployerRoutes(r *gin.RouterGroup, employerHandler *handler.EmployerHandler) {
	r.GET("/Profile", employerHandler.GetProfile)
	r.PUT("/Profile/Update", employerHandler.UpdateProfile)
	r.POST("/signup", employerHandler.Signup)
	r.POST("/login", employerHandler.Login)
}