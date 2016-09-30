/* Package chatapp implements the server and the client library for the chatapp application */
package chatapp

// import statements
import (
	"bufio"
	"net"
	"log"
	"strings"
	"fmt"
	"time"
)

/* Struct representing the chatapp client */
type Client struct {
	name string
	conn net.Conn // connection socket
	address chan *Message // primary message channel
	err chan *Message // channel for recieving errors
	server chan *Message // server's address
}

//TODO: Make Client implement Stringer interface
//TODO: Consider implementing error interface for some of this guiys

// NewClient creates a new client and starts two gorountine to manage it.
// One to communicate with server, and the other to communicate with the client interface
func NewClient(username string, conn net.Conn, serverAddr chan *Message) {

	// client addresses
	self := make(chan *Message)
	err := make(chan *Message)
	
	// client structure
	newClient := Client{username, conn, self, err, serverAddr}

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

// manages the communication between the client gorountine and the interface
// by monitoring it for new messages or commands from the user
func (client *Client) monitor() {
	for {
		// reads messages sent from the client interface
		msg, err := bufio.NewReader(client.conn).ReadString('\n')
		if err != nil {
			log.Println(err)

			//TODO: client disconnected, logout client... send logout message to server gorountine
			username := client.name
			logoutMessage := Message{title:"command", subject:"?logout", body:username, sender:client}
			client.server <- &logoutMessage
			break
		}

		// parses them into commands or messages
		message := client.parseMessage(msg)

		// sends commands to the server to process, using serverInAddr
		if message.title == "command" {
			client.server <- message	
		} else { // sends message to the appropriate chatroom
			fmt.Println("sends message to the appropriate chatroom")
		}		
	}
}

// parses the messages sent from the interface to the client
// returns a Message struct 
func (client *Client) parseMessage(msg string) *Message {
	msgStr := Message{sender: client}
	msgClean := strings.TrimSpace(msg)

	if strings.HasPrefix(msg, "?") {
		content := strings.SplitN(msgClean, " ", -1) //TODO: Test for double spaces //BUG
	
		msgStr.title = "command"
		msgStr.subject = strings.TrimSpace(content[0])
		
		if len(content) > 1 {
			msgStr.body = strings.TrimSpace(content[1])	
		}
	} else {
		msgStr.title = "message"
		msgStr.subject = msgClean		
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
			_, err := client.conn.Write([]byte(message.String()))
			if err != nil {
				log.Println(err)
			}
		
		case message := <- client.err:
			_, err := client.conn.Write([]byte(message.String()))

			if err != nil {
				log.Println()
			}
		}		
	}
}