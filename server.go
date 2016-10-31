package chatapp

// import statements
import (
	"bufio"
	"net"
	"log"
	"fmt"
	"strings"
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
	clients map[string]*Client // keep records of connected clients
	rooms map[string]*Chatroom // keep record of created chatrooms
	size int // tracks the number of connected clients
}

// Starts a tcp server on port p
func StartServer(p string) {
	
	// server goroutine communication channel
	s := make(chan *Message)
	
	// server object
	server := Server{port: p, protocol: "tcp", address: s, clients: make(map[string]*Client, MAXCLIENT), rooms: make(map[string]*Chatroom)}

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
		case "?logout": server.logout(msg)
		case "?help": server.help(msg)
		default: log.Println("Unknown command: " + msg.String())
		}		
	}

}

// Proxy listens and handle each client connection in new gorountine
// multiple connections may be served concurrently
func (s *Server) proxy() {

	l, err := net.Listen(s.protocol, s.port)
	if err != nil {
		log.Fatal(err) // fatal server error
	}

	log.Println("Server running ...")
	defer l.Close()
	
	// handle connections
	for {
		// wait for connection
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go s.connect(conn)
	}
}

func (s *Server) connect(conn net.Conn) {

	// recieve first message - the client username
	username, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		log.Println(err)		
	} else {
		// trim name before sending
		uname := strings.TrimSpace(username);

		// launch a new goroutine to create and manage client
		go NewClient(uname, conn, s.address)		
	}
}

// process login of the client to the server
func (s *Server) login(msg *Message) {

	// confirm that the client limit is not exceeded
	newClient := msg.sender
	if s.size >= MAXCLIENT {
		// send login failure message to interface
		newClient.err <- &Message{subject:"Error",body:"Current client count exceeded maximum. The server cannot login anymore client"}
		return
	}

	clientname := newClient.name
	if clientname == "" {
		newClient.err <- &Message{subject:"Error:",body:"Username cannot be empty. Exit and login with new username"}
		return
	}
	
	// confirm that user is not already logged-in
	if _ , ok := s.clients[clientname]; ok {
		// send login failure message to interface
		// TODO: create constants for all error messages
		newClient.err <- &Message{subject:"Error:",body:"Username already used. Exit and login with new username"}
		return
	} else {
		// add client to loggedin users
		s.clients[clientname] = newClient
		s.size++

		// send success message
		newClient.address <- &Message{subject:"login successful"}

		// send usage instructions to user
		s.help(msg)
		log.Print(clientname + " logged in")
	}
}

func (s *Server) createRoom(msg *Message) {

	client := msg.sender

	// // confirm that user is already logged-in
	// if _ , ok := s.clients[newClient.name]; ok {

	// check if chatroom exist
	roomname := strings.TrimSpace(msg.body);
	if roomname == "" {
		client.err <- &Message{subject:"Error:", body:fmt.Sprintf("Chatroom cannot have an empty name. Use a different name")}
		log.Println("Create chatroom failed")
		return
	}

	if _ , ok := s.rooms[roomname]; ok {
		// send login failure message to interface
		// TODO: create constants for all error messages
		client.err <- &Message{subject:"Error:", body:fmt.Sprintf("Chatroom already exist. Use: \"?join %s\" (without the quotes) to join room.", roomname)}
		log.Println("Create chatroom failed")	
	
	} else {
		// create new chatroom
		s.rooms[roomname] = NewChatroom(roomname)

		// send success message and usage instructions to user
		// log.Print(roomname)
		client.address <- &Message{body:fmt.Sprintf("Chatroom created. Use: \"?join %s\" (without the quotes) to join room.", roomname)}
		log.Printf("Created %s chatroom", roomname)
	}
}

func (s *Server) joinRoom(msg *Message) {

	roomname := strings.TrimSpace(msg.body)
	client := msg.sender
	if room, ok := s.rooms[roomname]; ok {		
		// join chatroom
		// check if room exists ---- done!
		// send a message to room address with client details
		room.commChannel <- &Message{subject: "join", body:client.name, sender:client}	
	} else {
		// send login failure message to interface
		// TODO: create constants for all error messages
		client.err <- &Message{subject:"Error:", body:fmt.Sprintf("Chatroom does not exist. Use: \"?create %s\" (without the quotes) to create room.", roomname)}
		log.Print("Join chatroom failed")
	}	
}


func (s *Server) leaveRoom(client string) {
	log.Println("leaveRoom")
}

// send list of available rooms
func (s *Server) listRooms(msg *Message) {
	client := msg.sender
	client.address <- &Message{subject:"Available rooms"}
	for i, _ := range s.rooms {
		client.address <- &Message{body:i}		
	}
}

// send usage instructions to client
func (s *Server) help(msg *Message) {

	client := msg.sender

	client.address <- &Message{subject:"Usage instructions"}
	client.address <- &Message{body:"?create AbC -> creates a chat room and set name to AbC"}		
	client.address <- &Message{body:"?list -> list the existing rooms"}
	client.address <- &Message{body:"?join AbC -> join chatroom AbC"}
	client.address <- &Message{body:"?leave AbC -> leave chatroom AbC"}
	client.address <- &Message{body:"?logout -> disconnect"}
	client.address <- &Message{body:"?help -> usage instructions"}
}

// remove user from connected chatrooms and connected client list
func (s *Server) logout(msg *Message) {

	client := msg.sender
	clientname := client.name

	// TODO: leave rooms already joined

	// remove name from client list
	client.err <- &Message{subject:"Exit:", body:"Session has been disconnected. Close window."}
	delete(s.clients, clientname)
	
	log.Println(clientname, "disconnected")
}
