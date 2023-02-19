package main

import (
	"fmt"
	"io"
	"os"

	"github.com/gin-gonic/gin"
)

func setupStorageDir() {
	storagePath := "./storage"
	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		os.Mkdir(storagePath, os.ModePerm)
		os.Chmod(storagePath, os.ModePerm)

	}
}

func setupLogFile() {
	logFilePath := "server.log"
	logFile, err := os.Create(logFilePath)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)

	}
	gin.DefaultWriter = io.MultiWriter(logFile, os.Stdout)

}

func setupServer() {
	//gin.SetMode(gin.ReleaseMode)

	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "admin.json")
	setupLogFile()
	setupStorageDir()
	// setupResultsDir()

}

func envPort() string {
	envPort := os.Getenv("PORT")
	if envPort == "" {
		envPort = "8080"
	}
	return ":" + envPort

}
