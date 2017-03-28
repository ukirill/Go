package main

import (
	q "clientqueue"
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"regexp"
)

const MAX_CLIENTS = 5
const MAX_QUEUE = 30
const STATE = "$STATE"
const ERROR = "$ERROR"

var clients = make(map[*websocket.Conn]string)       // all connected clients
var onlineClients = make(map[*websocket.Conn]string) // allowed clients
var queueClients = q.NewQueue(MAX_QUEUE)             // delayed clients order
var broadcast = make(chan Message)                   // broadcast channel
var upgrader = websocket.Upgrader{}

type Message struct {
	Username string `json:"username"`
	Text     string `json:"text"`
	Service  bool   `json:"service"`
	Time     string `json:"time"`
}

type State struct {
	Username string `json:"username"`
	Status   int    `json:"status"`
}

func main() {
	// Create a simple file server
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	// Configure websocket route
	http.HandleFunc("/ws", handleConnections)
	// Start listening for incoming chat messages
	go handleMessages()
	// Start UserList update service
	//go updateListOfUsersOnTime()

	log.Println("http server started on :8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
} //nothing to touch

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()
	// Register our new client
	clients[ws] = ""

	for {
		var msg Message
		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error in readmsg: %v", err)
			deleteClient(ws)
			break
		}
		// Send the newly received message to the broadcast channel or react
		// NEED PARSING OF SERVICE MESSAGES HERE
		if msg.Service {
			parseServiceMessage(&msg, ws)
		} else {
			broadcast <- msg
		}
	}
}

func handleMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast
		// Send it out to every client that is currently connected
		for client := range onlineClients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error writemsg: %v", err)
				client.Close()
				deleteClient(client)
			}
		}
	}
}

func sendStateMessage(ws *websocket.Conn, st *State) {
	var msg Message
	msg.Username = STATE
	msg.Service = true
	res, _ := json.Marshal(&st)
	msg.Text = string(res)
	msg.Time = ""
	ws.WriteJSON(msg) //Send new state
}

func sendErrorMessage(ws *websocket.Conn, st *State) {
	var msg Message
	msg.Username = ERROR
	msg.Service = true
	res, _ := json.Marshal(&st)
	msg.Text = string(res)
	msg.Time = ""
	ws.WriteJSON(msg) //Send new state
}

func checkLogin(s string) bool {
	var str = "[$,]"
	regex, _ := regexp.Compile(str)
	if s == "" {
		return false
	} else if regex.MatchString(s) {
		return false
	}
	return true
} //

func parseServiceMessage(msg *Message, ws *websocket.Conn) {
	if msg.Username == STATE {
		var st = State{}
		var cl = q.Client{}
		if json.Unmarshal([]byte(msg.Text), &st) == nil {
			if checkLogin(st.Username) {
				if len(onlineClients) < MAX_CLIENTS { //Go to chat
					onlineClients[ws] = st.Username
					sendStateMessage(ws, &st)
					refreshUsersList()
				} else if queueClients.Count < MAX_QUEUE { //Go to queue
					cl.WS = ws
					cl.Username = st.Username
					st.Status = 2
					queueClients.Add(&cl)
					sendStateMessage(ws, &st)
					//Send new state
				} else {
					st.Status = -1 //Everything full, drop
					sendStateMessage(ws, &st)
					ws.Close()
				}
			} else { //Login didnt pass check, error, initial state {"",0}
				st.Status = 0
				st.Username = "Login error, may be blank login"
				sendErrorMessage(ws, &st)
				st.Username = ""
				sendStateMessage(ws, &st)
			}
		}
	}
}

func addClientFromQueue() {
	if len(onlineClients) < MAX_CLIENTS {
		dc := queueClients.Next()
		var st State
		for {
			if dc != nil { //Queue not empty
				if dc.WS != nil { //Client in queue online
					clients[dc.WS] = dc.Username
					onlineClients[dc.WS] = dc.Username
					st.Username = dc.Username
					st.Status = 1
					sendStateMessage(dc.WS, &st)
					return
				} else { //Have disconnected
					dc = queueClients.Next()
				}
			} else { //Queue is empty
				return
			}
		}
	}
}

func deleteClient(client *websocket.Conn) {
	delete(clients, client)
	delete(onlineClients, client)
	addClientFromQueue()
	refreshUsersList()
}

func refreshUsersList() {
	var serviceMsg = Message{
		Username: "$USERLIST",
		Text:     makeListOfUsers(),
		Service:  true,
		Time:     "",
	}
	broadcast <- serviceMsg
}

func makeListOfUsers() string {
	var listOfUsers string = ""
	for _, username := range onlineClients {
		listOfUsers += username + ","
	}
	return listOfUsers
}
