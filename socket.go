package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var m = make(map[*websocket.Conn]bool)

func reader(conn *websocket.Conn) {
	defer func() {
		conn.Close()
		delete(m, conn)
	}()
	for {

		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		for c, _ := range m {

			c.WriteMessage(messageType, p)
		}

	}

}

func serveSocket(w http.ResponseWriter, r *http.Request) {

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
	}
	log.Println("Client connected")
	m[ws] = true
	fmt.Println(len(m)) //Prints number of active connections
	reader(ws)

}

func serveHome(w http.ResponseWriter, r *http.Request) {

	http.ServeFile(w, r, "home.html") //Register user name form

}

func registerUser(w http.ResponseWriter, r *http.Request) {
	log.Println("In register function")
	name := r.FormValue("userName")

	cookie := http.Cookie{Name: "chat-user-name", Value: name, Expires: time.Now().Add(60 * time.Minute), HttpOnly: true}

	http.SetCookie(w, &cookie)
	log.Println("Cookie set")
	http.Redirect(w, r, "/chat-room", http.StatusTemporaryRedirect)
}

func serveRoom(w http.ResponseWriter, r *http.Request) {
	if _, err := r.Cookie("chat-user-name"); err != nil {

		log.Println("No cookie found")
		http.Redirect(w, r, "/", 302)
	} else {

		http.ServeFile(w, r, "socket.html")
	}
}

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/", serveHome)
	router.HandleFunc("/ws", serveSocket)
	router.HandleFunc("/chat-room", serveRoom)
	router.HandleFunc("/register", registerUser)
	http.ListenAndServe(":8080", router)
}
