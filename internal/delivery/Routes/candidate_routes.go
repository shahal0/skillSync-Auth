package routes

import (
	"github.com/gin-gonic/gin"
	handler "skillsync-authservice/internal/delivery/handler"
)

func RegisterCandidateRoutes(r *gin.RouterGroup, candidateHandler *handler.CandidateHandler) {
	r.GET("/Profile", candidateHandler.GetProfile)
	r.POST("/AddSkills", candidateHandler.UpdateSkills)
	r.POST("/AddEducation", candidateHandler.UpdateEducation)
	r.PUT("/Profile/Update", candidateHandler.UpdateProfile)
	r.POST("signup", candidateHandler.Signup)
	r.POST("login", candidateHandler.Login)

}