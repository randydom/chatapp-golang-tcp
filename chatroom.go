package chatapp

import (
	"log"
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
	messages []*Message
	clientProxy *Client
}

// NewChatroom creates a chatroom and opens it in a new goroutine
func NewChatroom(roomname string) *Chatroom {

	// setup channels for communcation
	roomChan1 := make(chan *Message)
	roomChan2 := make(chan *Message)
	msgBox := make([]*Message, 0, MESSAGECAP)
	clientList := make(map[string] chan *Message)
	clientP := Client{}
	clientP.address = roomChan1

	// create room
	chatroom := Chatroom{name:roomname, clients:clientList, msgChannel: roomChan1, commChannel:roomChan2, messages:msgBox, clientProxy:&clientP}

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
				address := communique.sender.address

				if _, ok := room.clients[user]; ok {
					address <- &Message{body:fmt.Sprintf("You are already in this chatroom, use ?%s followed by space and the message to send to room", room.name), sender: room.clientProxy}
				} else {
					room.clients[user] = address
					responseMessage := Message{title:"admin", subject:room.name, body:fmt.Sprintf(": You have joined chatroom, use ?%s followed by space and the message to send to room", room.name), sender: room.clientProxy}
					address <- &responseMessage
					log.Printf("%s has joined %s chatroom", user, room.name)

					for _,v := range room.messages {
						address <- v
					}
				}
			} else {
					//TODO: handle later
					log.Println(communique)
			}			
		}
	}		
}