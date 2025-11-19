package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// В учебных проектах проще всего:
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func serveWs(hub *Hub, c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}

	client := NewClient(hub, conn)
	hub.register <- client

	go client.writePump()
	go client.readPump()
}

func main() {
	hub := NewHub()
	go hub.Run()

	r := gin.Default()

	// отдаём статику (наш фронт)
	r.Static("/static", "./web")

	// простая страница по корню
	r.GET("/", func(c *gin.Context) {
		c.File("./web/index.html")
	})

	// WebSocket endpoint
	r.GET("/ws", func(c *gin.Context) {
		serveWs(hub, c)
	})

	log.Println("server started on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
