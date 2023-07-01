package main

import (
	"main/logger"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// Create a new instance of the Gin engine
	r := gin.Default()

	// Get logger
	logger.InitStdoutLogger()
	test_logger, _ := logger.GetLogger("test-logger")

	// Define your routes
	r.GET("/health", func(c *gin.Context) {

		// connectionStr := "cefb0946-dbca-4b23-bc3c-9fc355a98436"
		// logger.SetConnectionString(connectionStr)
		logger.InitializeAppInsightsLogger(c)

		time.Sleep(10 * time.Second)

		c.JSON(200, gin.H{
			"message": "Hello, world!",
		})

		test_logger.Info("Hello app insights!")
	})

	// Run the server
	r.Run(":8080")
}
