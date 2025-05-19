package logger

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func InitLogger() {
	Log = logrus.New()
	Log.SetFormatter(&logrus.JSONFormatter{}) 
	Log.SetOutput(os.Stdout)                  
	Log.SetLevel(logrus.InfoLevel)           
}

func HandleError(c *gin.Context, statusCode int, err error, message string) {
	Log.WithError(err).Error(message)
	c.JSON(statusCode, gin.H{"error": message})
}
func InfoLog(c *gin.Context, message string) {
	Log.Info(message)
	c.JSON(200, gin.H{"message": message})
}
