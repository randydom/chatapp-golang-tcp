package chatapp

import (
	"log"
	"time"
	"fmt"
)

// constants
const MESSAGECAP = 1000

// Struct representing a Chatroom
type Chatroom struct {
	name string
	clients	map[string] chan *Message
	msgChannel chan *Message
	commChannel chan *Message
	server chan *Message // server's address
	messages []*Message
}

// NewChatroom creates a chatroom and opens it in a new goroutine
func NewChatroom(roomname string, serverAddr chan *Message) *Chatroom {

	// setup channels for communcation
	roomChan1 := make(chan *Message)
	roomChan2 := make(chan *Message)
	msgBox := make([]*Message, 0, MESSAGECAP)
	clientList := make(map[string] chan *Message)


	// create room
	chatroom := Chatroom{name:roomname, clients:clientList, msgChannel: roomChan1, commChannel:roomChan2, server:serverAddr, messages:msgBox}

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
			user := communique.body

			if communique.subject == "join" {
				address := communique.sender.address
				clientA := &Client{}
				clientA.address = room.msgChannel

				if _, ok := room.clients[user]; ok {
					address <- &Message{body:fmt.Sprintf("You are already in this chatroom, use ?%s followed by space and the message to send to room", room.name), sender: clientA}
				} else {
					room.clients[user] = address
					clientB := &Client{}
					clientB.address = room.commChannel

					responseMessageA := Message{title:"chatroomA", subject:room.name, body:fmt.Sprintf(": You have joined chatroom"), sender: clientA}
					responseMessageB := Message{title:"chatroomB", subject:room.name, body:fmt.Sprintf(": Use ?%s followed by space and the message to send to room", room.name), sender: clientB}

					address <- &responseMessageA
					address <- &responseMessageB

					log.Printf("%s has joined %s chatroom", user, room.name)

					for _,v := range room.messages {
						address <- v
					}
				}
			}
			if communique.title == "leave" {
				if _, ok := room.clients[user]; ok {
					delete(room.clients, user)
					log.Printf("%s has left %s chatroom", user, room.name)
				}
			}					
		
		case <- time.After(time.Hour * 24 * 7):
			// destroy chatroom after 7 days
			for _,v := range room.clients {
				v <- &Message{title:"Expired", subject:room.name, body:fmt.Sprintf(": Chatroom has expired.")}
			}

			room.server <- &Message{subject:"?destroy", body:room.name}
			return
		}
	}		
}