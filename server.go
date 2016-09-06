package chatapp

// import statements
import (
	"bufio"
	"net"
	"log"
	"fmt"
)

const MAXCLIENT = 10 // maximum number of client to manage concurrently

// Struct representing message passed between channels
type Message struct {
	title string // form includes command, message
	subject string 
	body string
	sender *Client
}

// String method to make the Message struct implement a Stringer interface
func (m *Message) String() string {
	return fmt.Sprintf("%s %s\n", m.subject, m.body) 
}

/* Struct representing the chatapp server */
type Server struct{
	port string
	protocol string
	address chan *Message
	clients map[string]*Client
	rooms map[string]*Chatroom
	size int
}

// Starts a tcp server on port p
func StartServer(p string) {
	
	// server goroutine communication channel
	s := make(chan *Message)
	
	// server object
	server := Server{port: p, protocol: "tcp", address: s, clients: make(map[string]*Client), rooms: make(map[string]*Chatroom)}

	// spawn a proxy goroutine that manages the connections
	go server.proxy()

	// continue running
	for {
		msg := <- s

		// read message subject and call the appropriate method
		switch msg.subject {
		case "?login": server.login(msg)
		case "?list": server.listRooms(msg)
		case "?join": server.joinRoom(msg)
		case "?create": server.createRoom(msg)
		case "?leave": server.leaveRoom(msg.body)
		case "?logout": server.logout(msg.body)
		default: log.Println(msg)
		}		
	}

}

// Proxy listens and handle each client connection in new gorountine
// multiple connections may be served concurrently
func (s *Server) proxy() {

	l, err := net.Listen(s.protocol, s.port)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Server running...")
	defer l.Close()
	
	// handle connections
	for {
		// wait for connection
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		// recieve first message - the client username
		username, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		
		// launch a new goroutine to create and manage client
		go NewClient(username, conn, s.address)
	}
}


// process login of the client to the server
func (s *Server) login(msg *Message) int {

	// confirm that the client limit is not exceeded
	newClient := msg.sender
	if s.size >= MAXCLIENT {
		// send login failure message to interface
		newClient.err <- &Message{subject:"Error",body:"Current client count exceeded maximum. The server cannot login anymore client"}
		return 1
	}
	
	// confirm that user is not already logged-in
	if _ , ok := s.clients[newClient.name]; ok {
		// send login failure message to interface
		// TODO: create constants for all error messages
		newClient.err <- &Message{subject:"Error:",body:"Username already used. Exit and login with new username"}
		//TODO: login with anonymous and force to change immediately
	
	} else {
		// add client to loggedin users
		s.clients[newClient.name] = newClient
		s.size++

		// send success message and usage instructions to user
		newClient.address <- &Message{subject:"login successful"}
		newClient.address <- &Message{subject:"Usage instructions"}
		newClient.address <- &Message{body:"?create AbC -> creates a chat room and set name to AbC"}		
		newClient.address <- &Message{body:"?list -> list the existing rooms"}
		newClient.address <- &Message{body:"?join AbC -> join chatroom AbC"}
		newClient.address <- &Message{body:"?leave AbC -> leave chatroom AbC"}
		newClient.address <- &Message{body:"?logout -> disconnect"}
	}

	return 0
}

func (s *Server) createRoom(msg *Message) {
	
	// check if chatroom exist
	client := msg.sender
	roomname := msg.body

	if _ , ok := s.rooms[roomname]; ok {
		// send login failure message to interface
		// TODO: create constants for all error messages
		client.err <- &Message{subject:"Error:", body:fmt.Sprintf("Chatroom already exist. Use: join %s to join room.", roomname)}
		log.Println("Create chatroom failed")	
	
	} else {
		// create new chatroom
		s.rooms[roomname] = NewChatroom(roomname)

		// send success message and usage instructions to user
		log.Print(roomname)
		client.address <- &Message{body:fmt.Sprintf("Chatroom created. Use: \"join %s\" to join room.", roomname)}
		log.Println("Chatroom created")
	}
}

func (s *Server) joinRoom(msg *Message) {
	// check if chatroom exist
	roomname := msg.body
	client := msg.sender
	if _ , ok := s.rooms[roomname]; ok {
		
		// join chatroom
		// if response, ok := s.chatroom[roomname].JoinRoom(client); ok {
			// send success message and usage instructions to user
			client.address <- &Message{subject: "Success", body:fmt.Sprintf("Joined room. Use: ?%s followed by your message to send to room.", roomname)}
			log.Println("Joined room")
		
		// } else {
		// 	client.address <- &Message{subject:"Error", body: response}
		// 	log.Println("Join room failed")
		// }
	
	} else {
		// send login failure message to interface
		// TODO: create constants for all error messages
		client.err <- &Message{subject:"Error:", body:fmt.Sprintf("Chatroom already exist. Use: join %s to join room.", roomname)}
		log.Println("Create chatroom failed")	

	}	
}

func (s *Server) leaveRoom(client string) {
	log.Println("leaveRoom")
}

func (s *Server) listRooms(msg *Message) {
	log.Println("listRooms")
}

func (s *Server) logout(client string) {
	log.Println("leaveRoom")
}
