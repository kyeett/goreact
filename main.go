package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/satori/go.uuid"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/websocket"
)

var hub *Hub

func main() {
	hub = newHub()
	go hub.run()

	// Set the router as the default one shipped with Gin
	router := gin.Default()

	router.GET("/ws", func(c *gin.Context) {
		handler := websocket.Handler(func(conn *websocket.Conn) {

			clientCh := make(chan bool)
			id := uuid.NewV4().String()[0:10]
			hub.clients[id] = clientCh

			encoder := json.NewEncoder(conn)
			t := time.NewTicker(1 * time.Second)
			for {

				select {
				case <-t.C:
					fmt.Println("Time out received")
				case <-clientCh:
					fmt.Println("Update received")
				}
				if err := encoder.Encode(jokes); err != nil {
					fmt.Println("encode.Encode error: ", err)
				}
			}
		})
		handler.ServeHTTP(c.Writer, c.Request)
	})

	// Serve frontend static files
	router.Use(static.Serve("/", static.LocalFile("./views", true)))

	// Setup route group for the API
	api := router.Group("/api")
	{
		api.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "pong",
			})
		})
	}

	api.GET("/jokes", JokeHandler)
	api.POST("/jokes/like/:jokeID", LikeJoke)

	// Start and run the server
	router.Run(":3000")
}

type Joke struct {
	ID    int    `json:"id" binding:"required"`
	Likes int    `json:"likes"`
	Joke  string `json:"joke" binrding:"required"`
}

// We'll create a list of jokes
var jokes = []Joke{
	Joke{1, 0, "Did you hear about the restaurant on the moon? Great food, no atmosphere."},
	Joke{2, 0, "What do you call a fake noodle? An Impasta."},
	Joke{3, 0, "How many apples grow on a tree? All of them."},
	Joke{4, 0, "Want to hear a joke about paper? Nevermind it's tearable."},
	Joke{5, 0, "I just watched a program about beavers. It was the best dam program I've ever seen."},
	Joke{6, 0, "Why did the coffee file a police report? It got mugged."},
	Joke{7, 0, "How does a penguin build it's house? Igloos it together."},
}

// JokeHandler retrieve a list of available jokes
func JokeHandler(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, jokes)
}

// LikeJoke increments the likes of a particular joke
func LikeJoke(c *gin.Context) {

	jokeid, err := strconv.Atoi(c.Param("jokeID"))
	fmt.Println(err)
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	for i, j := range jokes {
		if j.ID == jokeid {
			jokes[i].Likes++
			hub.broadcast <- []byte("") //@TODO clean up
		}
	}
	c.JSON(http.StatusOK, &jokes)
}
