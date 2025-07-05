package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type ChatMessage struct {
	ID        int       `json:"id"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Read      bool      `json:"read"`
}

type App struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	queue     amqp.Queue
	messages  []ChatMessage
	users     []User
	messageID int
	userID    int
}

func NewApp() *App {
	return &App{
		messages:  make([]ChatMessage, 0),
		users:     make([]User, 0),
		messageID: 0,
		userID:    0,
	}
}

func (app *App) initUsers() {
	// Buat 2 user default untuk demo
	app.users = []User{
		{ID: 1, Username: "alice", Password: "password123"},
		{ID: 2, Username: "bob", Password: "password123"},
	}
	app.userID = 2
}

func (app *App) connectRabbitMQ() error {
	var err error

	app.conn, err = amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	app.channel, err = app.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %v", err)
	}

	app.queue, err = app.channel.QueueDeclare(
		"chat_messages", // name
		false,           // durable
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

func (app *App) publishMessage(from, to, content string) error {
	app.messageID++
	msg := ChatMessage{
		ID:        app.messageID,
		From:      from,
		To:        to,
		Content:   content,
		Timestamp: time.Now(),
		Read:      false,
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
			var msg ChatMessage
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				log.Printf("Error unmarshaling message: %v", err)
				continue
			}

			log.Printf("Message delivered: %s -> %s: %s", msg.From, msg.To, msg.Content)
		}
	}()
}

func (app *App) authenticateUser(username, password string) *User {
	for _, user := range app.users {
		if user.Username == username && user.Password == password {
			return &user
		}
	}
	return nil
}

func (app *App) requireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		username := session.Get("username")
		if username == nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		c.Set("username", username)
		c.Next()
	}
}

func (app *App) getOtherUser(currentUser string) string {
	for _, user := range app.users {
		if user.Username != currentUser {
			return user.Username
		}
	}
	return ""
}

func (app *App) getMessagesForUser(username string) []ChatMessage {
	var userMessages []ChatMessage
	for _, msg := range app.messages {
		if msg.From == username || msg.To == username {
			userMessages = append(userMessages, msg)
		}
	}
	return userMessages
}

func (app *App) setupRoutes() *gin.Engine {
	r := gin.Default()

	// Session middleware
	store := cookie.NewStore([]byte("secret-key-change-this"))
	r.Use(sessions.Sessions("session", store))

	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./static")

	// Public routes
	r.GET("/login", app.loginPage)
	r.POST("/login", app.loginHandler)
	r.GET("/logout", app.logoutHandler)

	// Protected routes
	authorized := r.Group("/")
	authorized.Use(app.requireAuth())
	{
		authorized.GET("/", app.chatPage)
		authorized.POST("/send", app.sendMessage)
		authorized.GET("/messages", app.getMessages)
		authorized.POST("/mark-read/:id", app.markAsRead)
	}

	return r
}

func (app *App) loginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"title": "Login - Chat App",
	})
}

func (app *App) loginHandler(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	user := app.authenticateUser(username, password)
	if user == nil {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"title": "Login - Chat App",
			"error": "Username atau password salah",
		})
		return
	}

	session := sessions.Default(c)
	session.Set("username", user.Username)
	session.Set("userID", user.ID)
	session.Save()

	c.Redirect(http.StatusFound, "/")
}

func (app *App) logoutHandler(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.Redirect(http.StatusFound, "/login")
}

func (app *App) chatPage(c *gin.Context) {
	username := c.MustGet("username").(string)
	otherUser := app.getOtherUser(username)

	c.HTML(http.StatusOK, "chat.html", gin.H{
		"title":     "Chat App",
		"username":  username,
		"otherUser": otherUser,
	})
}

func (app *App) sendMessage(c *gin.Context) {
	username := c.MustGet("username").(string)
	content := c.PostForm("content")
	to := c.PostForm("to")

	if content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Content cannot be empty"})
		return
	}

	if to == "" {
		to = app.getOtherUser(username)
	}

	err := app.publishMessage(username, to, content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}

	// Return updated messages
	messages := app.getMessagesForUser(username)
	c.HTML(http.StatusOK, "messages.html", gin.H{
		"messages":    messages,
		"currentUser": username,
	})
}

func (app *App) getMessages(c *gin.Context) {
	username := c.MustGet("username").(string)
	messages := app.getMessagesForUser(username)

	c.HTML(http.StatusOK, "messages.html", gin.H{
		"messages":    messages,
		"currentUser": username,
	})
}

func (app *App) markAsRead(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Mark message as read
	for i, msg := range app.messages {
		if msg.ID == id {
			app.messages[i].Read = true
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
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

	// Initialize users
	app.initUsers()

	// Connect to RabbitMQ
	err := app.connectRabbitMQ()
	if err != nil {
		log.Fatal(err)
	}

	// Start consuming messages
	app.consumeMessages()

	// Setup routes
	r := app.setupRoutes()

	fmt.Println("Chat app running on http://localhost:4300")
	fmt.Println("Users: alice/password123 dan bob/password123")
	log.Fatal(r.Run(":4300"))
}
