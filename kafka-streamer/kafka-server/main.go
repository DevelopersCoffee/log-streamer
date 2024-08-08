package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
)

var (
	producer sarama.SyncProducer
	consumer sarama.Consumer
	admin    sarama.ClusterAdmin
)

func main() {
	r := gin.Default()

	// Serve static files
	r.Static("/static", "./static")

	// Produce message endpoint
	r.GET("/api/produce/:topic/:message", produceMessage)

	// Delete topic endpoint
	r.DELETE("/api/delete/:topic", deleteTopic)

	// List topics endpoint
	r.GET("/api/topics", listTopics)

	// Server-Sent Events endpoint
	r.GET("/api/events/:topic/:limit", sseEndpoint)

	// Kafka setup
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	var err error
	producer, err = sarama.NewSyncProducer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatalf("Failed to start Sarama producer: %v", err)
	}
	defer producer.Close()

	consumer, err = sarama.NewConsumer([]string{"localhost:9092"}, nil)
	if err != nil {
		log.Fatalf("Failed to start Sarama consumer: %v", err)
	}
	defer consumer.Close()

	admin, err = sarama.NewClusterAdmin([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatalf("Failed to create Kafka admin client: %v", err)
	}
	defer admin.Close()

	r.Run(":8080")
}

func produceMessage(c *gin.Context) {
	topic := c.Param("topic")
	message := c.Param("message")

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}

	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "message produced",
		"partition": partition,
		"offset":    offset,
	})
}

func deleteTopic(c *gin.Context) {
	topic := c.Param("topic")

	err := admin.DeleteTopic(topic)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "topic deleted",
	})
}

func listTopics(c *gin.Context) {
	topics, err := admin.ListTopics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	topicNames := make([]string, 0, len(topics))
	for topic := range topics {
		topicNames = append(topicNames, topic)
	}

	c.JSON(http.StatusOK, gin.H{
		"topics": topicNames,
	})
}

func sseEndpoint(c *gin.Context) {
	topic := c.Param("topic")
	limitStr := c.Param("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50 // default limit
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		http.Error(c.Writer, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest-int64(limit))
	if err != nil {
		http.Error(c.Writer, fmt.Sprintf("Failed to start consumer: %v", err), http.StatusInternalServerError)
		return
	}
	defer partitionConsumer.Close()

	for {
		select {
		case msg := <-partitionConsumer.Messages():
			fmt.Fprintf(c.Writer, "data: %s\n\n", string(msg.Value))
			flusher.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}
