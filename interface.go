package chatapp

import (
	"log"
	"net"
	"bufio"
	"fmt"
	"os"
)

func StartClient(port string) {

	// dail the server
	conn, err := net.Dial("tcp", port)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Welcome.")
	fmt.Print("Login with your username: ")

	go Listen(conn)

	// listen for message on the standard input and echo it to the connection socket
	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(conn)

	for { 
    	// read input from stdin
    	input, err := reader.ReadString('\n')
		if err != nil {
			log.Println(err)
		}
    
    	// send to tcp socket
    	_, err = writer.WriteString(input)
		if err != nil {
			log.Println(err)
		}

		// always flush
		err = writer.Flush()
		if err != nil {
			log.Println(err)
		}
	}	
}

// listen for message on the connection socket and echo it to the standard output
func Listen(conn net.Conn) {
	
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(os.Stdout)

	for { 
		
		msg, err := reader.ReadString('\n')
	    if err != nil {
			log.Fatal(err)
		}

		_, err = writer.WriteString(msg)
	    if err != nil {
			log.Fatal(err)
		}

		err = writer.Flush()
	    if err != nil {
			log.Fatal(err)
		}
	}
}