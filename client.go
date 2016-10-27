/* Package chatapp implements the server and the client library for the chatapp application */
package chatapp

// import statements
import (
	"bufio"
	"net"
	"log"
	"strings"
	"time"
	"fmt"
)


var FACE = [5]string{"?list","?join","?create","?leave", "?logout"} //server api

/* Struct representing the chatapp client */
type Client struct {
	name string
	conn net.Conn // connection socket
	address chan *Message // primary message channel
	err chan *Message // channel for recieving errors
	server chan *Message // server's address
	rooms map[string] chan *Message //keep record of connected chatrooms address
}

//TODO: Make Client implement Stringer interface
//TODO: Consider implementing error interface for some of this guiys

// NewClient creates a new client and starts two gorountine to manage it.
// One to communicate with server, and the other to communicate with the client interface
func NewClient(username string, conn net.Conn, serverAddr chan *Message) {

	// client addresses
	self := make(chan *Message)
	err := make(chan *Message)
	rooms := make(map[string] chan *Message)
	
	// client structure
	newClient := Client{username, conn, self, err, serverAddr, rooms}

	// send login message to server gorountine
	loginMessage := Message{title:"command", subject:"?login", body:username, sender:&newClient}
	serverAddr <- &loginMessage

	// wait for confirmation
	select {
	case success := <- self:
		_, err := newClient.conn.Write([]byte(success.String()))

		if err != nil {
			log.Println(err)
		}
		
		// handle the session in a new gorountine.
		// one to listen on the client channel, the other to listen on the socket interface
		go newClient.listen()
		go newClient.monitor()
	
	case failure := <- err:
		_, err := newClient.conn.Write([]byte(failure.String()))

		if err != nil {
			log.Println(err)
		}

	case <- time.After(time.Millisecond * 1500):
		_, err := newClient.conn.Write([]byte("timed out"))

		if err != nil {
			log.Println(err)
		}
	}
}

func (client *Client) messageRoom(msg *Message) {

	// details
	room := msg.subject

	// locate room address
	if rmAddress, ok := client.rooms[room]; ok {		
		t := time.Now()
		info := fmt.Sprintf("[%v](%d-%d-%v %d:%d) %v: %v", strings.TrimPrefix(room, "?"), t.Day(), t.Month(), t.Year(), t.Hour(), t.Minute(), client.name, msg.body)
		// info := fmt.Sprintf("(%s) %s: %s", time.Now().Format(""), client.name, msg.body)

		// send message to address
		rmAddress <- &Message{title:"broadcast", subject:info}
	} else {
		client.err <- &Message{subject:"Error:", body:fmt.Sprintf("Either you are not a member of this room or the room does not exist. Use: \"?list\" (without the quotes) to list available rooms.")}
		fmt.Println("error")
	}
}

func (client *Client) logout() {
	//TODO: client disconnected, logout client... send logout message to server gorountine
	username := client.name
	logoutMessage := Message{title:"command", subject:"?logout", body:username, sender:client}
	client.server <- &logoutMessage
}

// manages the communication between the client gorountine and the interface
// by monitoring it for new messages or commands from the user
func (client *Client) monitor() {
	for {
		// reads messages sent from the client interface
		msg, err := bufio.NewReader(client.conn).ReadString('\n')
		if err != nil {
			log.Println(err)
			client.logout()
			break
		}

		// parses them into commands or messages
		msgClean := strings.TrimSpace(msg)
		if len(msgClean) == 0 {
			continue
		}

		message := client.parseMessage(msgClean)

		// sends commands to the server to process, using serverInAddr
		if message.title == "command" {
			client.server <- message
			if message.subject == "?logout" {
				break
			}
		} else { // sends message to the appropriate chatroom
			client.messageRoom(message)
		}		
	}
}

// parses the messages sent from the interface to the client
// returns a Message struct 
func (client *Client) parseMessage(msgClean string) *Message {
	
	msgStr := Message{sender: client}
	content := strings.Fields(msgClean)
	for _, b := range FACE {
		if b == content[0] {
			msgStr.title = "command"
			msgStr.subject = content[0]
			
			if len(content) > 1 {
				msgStr.body = strings.TrimSpace(content[1])	
			}
			return &msgStr
		}
	}

	msgStr.title = "message"
	msgStr.subject = content[0]
	if len(content) > 1 {
		msgStr.body = strings.Join(content[1:], " ")
	}

	return &msgStr
}

// manages the communication between the client gorountine and the server
// by listening on the designated channel
// recieved messages broadcasted by chatroom and responses from the server
func (client *Client) listen() {
	for {
		select {
		case message := <- client.address:

 			if message.title == "admin" {
 				// add chatroom address to client records
 				roomname := "?"+message.subject
 				client.rooms[roomname] = message.sender.address
 			}
			_, err := client.conn.Write([]byte(message.String()))
			if err != nil {
				log.Println(err)
			}
		
		case message := <- client.err:
			_, err := client.conn.Write([]byte(message.String()))

			if err != nil {
				log.Println(err)
			}

			if message.subject == "Exit:"  {
				break
 			}
		}		
	}
}