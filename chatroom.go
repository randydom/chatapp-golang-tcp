package chatapp

import (
	"log"
)

// constants
const MESSAGECAP = 1000

// Struct representing a Chatroom
type Chatroom struct {
	name string
	clients	map[string] chan *Message
	msgChannel chan *Message
	commChannel chan *Message
	messages []*Message
}

// NewChatroom creates a chatroom and opens it in a new goroutine
func NewChatroom(roomname string) *Chatroom {

	// setup channels for communcation
	roomChan1 := make(chan *Message)
	roomChan2 := make(chan *Message)
	msgBox := make([]*Message, 0, MESSAGECAP)
	// clientList := make([] string, 0, MAXCLIENT)
	clientList := make(map[string] chan *Message)

	// create room
	chatroom := Chatroom{name:roomname, clients:clientList, msgChannel: roomChan1, commChannel:roomChan2, messages:msgBox}

	// launch room in its open goroutine
	go chatroom.open()

	return &chatroom
}

// manages the operation of the chatroom
func (room *Chatroom) open() {

	for {
		// recieve a message on the room channel
		select {
		case message := <- room.msgChannel:
			// save in message archive
			room.messages = append(room.messages, message)

			// broadcast it on the message channel of connected clients
			for _,v := range room.clients {
				v <- message
			}						
		case communique := <- room.commChannel:
			if communique.subject == "join" {
				user := communique.body
				if _, ok := room.clients[user]; ok {
					log.Printf("User already in room, use ?%s followed by space and the message to send to room", room.name)
				} else {
					room.clients[user] = make(chan *Message)
					log.Printf("User has joined room, use ?%s followed by space and the message to send to room", room.name)
				}
			} else {
				log.Println(communique)
			}			
		}
	}		
}