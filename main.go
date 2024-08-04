package main

import (
	"bufio"
	"embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

//go:embed static/*
var staticFiles embed.FS

func main() {
	router := gin.Default()

	// Serve the index.html file for the root URL
	router.GET("/", func(c *gin.Context) {
		data, err := staticFiles.ReadFile("static/index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Error reading index.html")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	// Serve static files
	router.GET("/static/*filepath", func(c *gin.Context) {
		filePath := "static/static" + c.Param("filepath")
		file, err := staticFiles.Open(filePath)
		if err != nil {
			c.String(http.StatusNotFound, "File not found")
			return
		}
		defer file.Close()

		stat, err := file.Stat()
		if err != nil {
			c.String(http.StatusInternalServerError, "Error getting file info")
			return
		}

		http.ServeContent(c.Writer, c.Request, filePath, stat.ModTime(), file.(io.ReadSeeker))
	})

	// Serve log files for streaming
	router.GET("/api/logs/:filename", func(c *gin.Context) {
		filename := c.Param("filename")
		if !strings.HasSuffix(filename, ".log") {
			c.String(http.StatusBadRequest, "Invalid file type")
			return
		}

		filePath := filepath.Join("/tmp/local", filename)
		file, err := os.Open(filePath)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error opening file")
			return
		}
		defer file.Close()

		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")

		reader := bufio.NewReader(file)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					c.String(http.StatusInternalServerError, "Error reading file")
				}
				break
			}
			fmt.Fprintf(c.Writer, "data: %s\n\n", line)
			c.Writer.Flush()
			time.Sleep(1 * time.Second) // Simulate real-time log streaming
		}
	})

	// List log files
	router.GET("/api/files", func(c *gin.Context) {
		files, err := os.ReadDir("/tmp/local")
		if err != nil {
			c.String(http.StatusInternalServerError, "Error reading directory")
			return
		}

		var logFiles []string
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".log") {
				logFiles = append(logFiles, file.Name())
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"files": logFiles,
		})
	})

	router.Run(":8080")
}
