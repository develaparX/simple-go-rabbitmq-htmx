package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
)

type Message struct {
	ID        int       `json:"id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"`
}

type App struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	queue     amqp.Queue
	messages  []Message
	messageID int
}

func NewApp() *App {
	return &App{
		messages:  make([]Message, 0),
		messageID: 0,
	}
}

func (app *App) connectRabbitMQ() error {
	var err error

	// Koneksi ke RabbitMQ
	app.conn, err = amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	app.channel, err = app.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %v", err)
	}

	// Declare queue
	app.queue, err = app.channel.QueueDeclare(
		"message_queue", // name
		true,            // durable
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %v", err)
	}

	return nil
}

func (app *App) publishMessage(content string) error {
	app.messageID++
	msg := Message{
		ID:        app.messageID,
		Content:   content,
		Timestamp: time.Now(),
		Status:    "sent",
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	err = app.channel.Publish(
		"",             // exchange
		app.queue.Name, // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})

	if err != nil {
		return err
	}

	app.messages = append(app.messages, msg)
	return nil
}

func (app *App) consumeMessages() {
	msgs, err := app.channel.Consume(
		app.queue.Name, // queue
		"",             // consumer
		true,           // auto-ack
		false,          // exclusive
		false,          // no-local
		false,          // no-wait
		nil,            // args
	)
	if err != nil {
		log.Printf("Failed to register consumer: %v", err)
		return
	}

	go func() {
		for d := range msgs {
			var msg Message
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				log.Printf("Error unmarshaling message: %v", err)
				continue
			}

			msg.Status = "received"

			// Update message status in slice
			for i, m := range app.messages {
				if m.ID == msg.ID {
					app.messages[i] = msg
					break
				}
			}

			log.Printf("Received message: %s", msg.Content)
		}
	}()
}

func (app *App) setupRoutes() *gin.Engine {
	r := gin.Default()

	// Serve static files
	r.Static("/static", "./static")

	// HTML template
	r.LoadHTMLGlob("templates/*")

	// Routes
	r.GET("/", app.homePage)
	r.POST("/send", app.sendMessage)
	r.GET("/messages", app.getMessages)
	r.DELETE("/messages/:id", app.deleteMessage)

	return r
}

func (app *App) homePage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "Golang Gin + RabbitMQ + HTMX",
	})
}

func (app *App) sendMessage(c *gin.Context) {
	content := c.PostForm("content")
	if content == "" {
		c.String(http.StatusBadRequest, "Content cannot be empty")
		return
	}

	err := app.publishMessage(content)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to send message: %v", err)
		return
	}

	// Return updated messages list
	c.HTML(http.StatusOK, "messages.html", gin.H{
		"messages": app.messages,
	})
}

func (app *App) getMessages(c *gin.Context) {
	c.HTML(http.StatusOK, "messages.html", gin.H{
		"messages": app.messages,
	})
}

func (app *App) deleteMessage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid ID")
		return
	}

	// Remove message from slice
	for i, msg := range app.messages {
		if msg.ID == id {
			app.messages = append(app.messages[:i], app.messages[i+1:]...)
			break
		}
	}

	c.HTML(http.StatusOK, "messages.html", gin.H{
		"messages": app.messages,
	})
}

func (app *App) close() {
	if app.channel != nil {
		app.channel.Close()
	}
	if app.conn != nil {
		app.conn.Close()
	}
}

func main() {
	app := NewApp()
	defer app.close()

	// Connect to RabbitMQ
	err := app.connectRabbitMQ()
	if err != nil {
		log.Fatal(err)
	}

	// Start consuming messages
	app.consumeMessages()

	// Setup routes
	r := app.setupRoutes()

	fmt.Println("Server running on http://localhost:4300")
	log.Fatal(r.Run(":4300"))
}
