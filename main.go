package main

import (
	"embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
)

//go:embed static/*
var staticFiles embed.FS

// Client represents a single connection from a client.
type Client struct {
	logChannel chan string
	disconnect chan bool
}

// ClientManager handles all connected clients.
type ClientManager struct {
	clients map[string][]*Client
	mu      sync.Mutex
}

var clientManager = &ClientManager{
	clients: make(map[string][]*Client),
}

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

		client := &Client{
			logChannel: make(chan string),
			disconnect: make(chan bool),
		}

		clientManager.mu.Lock()
		clientManager.clients[filename] = append(clientManager.clients[filename], client)
		clientManager.mu.Unlock()

		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")

		c.Stream(func(w io.Writer) bool {
			select {
			case log := <-client.logChannel:
				fmt.Fprintf(w, "data: %s\n\n", log)
				return true
			case <-client.disconnect:
				return false
			}
		})

		client.disconnect <- true
		clientManager.mu.Lock()
		removeClient(filename, client)
		clientManager.mu.Unlock()
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

	// Download log file
	router.GET("/api/download/:filename", func(c *gin.Context) {
		filename := c.Param("filename")
		if !strings.HasSuffix(filename, ".log") {
			c.String(http.StatusBadRequest, "Invalid file type")
			return
		}

		filePath := filepath.Join("/tmp/local", filename)
		c.File(filePath)
	})

	go monitorLogFiles()

	router.Run(":8080")
}

func monitorLogFiles() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("ERROR", err)
		return
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					go tailLogFile(filepath.Base(event.Name))
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("ERROR", err)
			}
		}
	}()

	err = watcher.Add("/tmp/local")
	if err != nil {
		fmt.Println("ERROR", err)
	}
	// Add existing log files to the watcher
	files, err := os.ReadDir("/tmp/local")
	if err == nil {
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".log") {
				go tailLogFile(file.Name())
			}
		}
	}
	<-make(chan struct{}) // Keep the function running
}

func tailLogFile(filename string) {
	filePath := filepath.Join("/tmp/local", filename)
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	var lastReadPos int64

	for {
		stat, err := file.Stat()
		if err != nil {
			return
		}
		fileSize := stat.Size()
		if fileSize > lastReadPos {
			file.Seek(lastReadPos, io.SeekStart)
			buf := make([]byte, fileSize-lastReadPos)
			_, err := file.Read(buf)
			if err != nil && err != io.EOF {
				return
			}
			lastReadPos = fileSize

			lines := strings.Split(string(buf), "\n")
			for _, line := range lines {
				if line != "" {
					clientManager.mu.Lock()
					for _, client := range clientManager.clients[filename] {
						select {
						case client.logChannel <- line:
						case <-time.After(1 * time.Second):
							// Client not ready, skip
						}
					}
					clientManager.mu.Unlock()
				}
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func removeClient(filename string, client *Client) {
	clients := clientManager.clients[filename]
	for i, c := range clients {
		if c == client {
			clientManager.clients[filename] = append(clients[:i], clients[i+1:]...)
			break
		}
	}
}
