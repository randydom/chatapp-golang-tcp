package main

// import statements
import (
	"github.com/edeas123/chatapp-golang-tcp"
)

// main function
func main() {
	
	// retrieve some command line arguments
	port := "127.0.0.1:2000"
	
	// server goroutine communication channel
	chatapp.StartClient(port)
}