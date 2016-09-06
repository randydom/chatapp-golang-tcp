package main

// import statements
import (
	"github.com/edeas123/chatapp-golang-tcp"
)

// main function
func main() {
	
	// retrieve some command line arguments
	port := ":2000"
	
	// server goroutine communication channel
	chatapp.StartServer(port)

}