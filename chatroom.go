package main

import (
    "bufio" //used for I/O
    "flag" //used for flags when first running the program; run as client or server
    "fmt"  //used for scanning and printing
    "net" //used for server and client decleration
    "os"  //used for reading messages from a client
    "strings"  //used for message strings
)

type ClientManager struct {
    clients    map[*Client]bool
    broadcast  chan []byte
    register   chan *Client
    unregister chan *Client
}

type Client struct {
    socket net.Conn
    data   chan []byte
}

//Server allows clients to connect
func (manager *ClientManager) start() {
    for {
        select {
        case connection := <-manager.register:
            manager.clients[connection] = true
            fmt.Println("Added new connection!")
        case connection := <-manager.unregister:
            if _, ok := manager.clients[connection]; ok {
                close(connection.data)
                delete(manager.clients, connection)
                fmt.Println("A connection has terminated!")
            }
        case message := <-manager.broadcast:
            for connection := range manager.clients {
                select {
                case connection.data <- message:
                default:
                    close(connection.data)
                    delete(manager.clients, connection)
                }
            }
        }
    }
}

//Message from client to server, then server to client
func (manager *ClientManager) receive(client *Client) {
    for {
        message := make([]byte, 4096)
        length, err := client.socket.Read(message)
        //Checks client connection
        if err != nil { 
            manager.unregister <- client
            client.socket.Close()
            break
        }
        //If there is a message to be sent
        if length > 0 {
            fmt.Println("Clients message: " + string(message))
            manager.broadcast <- message
        }
    }
}

func (client *Client) receive() {
    for {
        message := make([]byte, 4096)
        length, err := client.socket.Read(message)
        //Checks client connection
        if err != nil {
            client.socket.Close()
            break
        }
        //If there is a message, prints it
        if length > 0 {
            fmt.Println("Recieved: " + string(message))
        }
    }
}

func (manager *ClientManager) send(client *Client) {
    //defer keyword delays close() funtion until the rest of the code finishs
    defer client.socket.Close()
    //Infinite for loop since no conditons
    for {
        //Switch statement
        select {
        case message, ok := <-client.data:
            if !ok {
                return
            }
            client.socket.Write(message)
        }
    }
}

//Creates Server
func startServer() {
    fmt.Println("Starting server...")
    //net.Listen allows for a client to connect with correct perameters
    listener, error := net.Listen("tcp", ":1")
    if error != nil {
        fmt.Println(error)
    }
    manager := ClientManager{
        clients:    make(map[*Client]bool),
        broadcast:  make(chan []byte),
        register:   make(chan *Client),
        unregister: make(chan *Client),
    }
    go manager.start()
    
    //Infinite for loop; since no condition, it is equivlant to "true"
    for {
        connection, _ := listener.Accept()
        if error != nil {
            fmt.Println(error)
        }
        client := &Client{socket: connection, data: make(chan []byte)}
        manager.register <- client
        go manager.receive(client)
        go manager.send(client)
    }
}

//Creates Client
func startClient() {
    fmt.Println("Starting client...")
    
    //net.Dial connects to a server of given perameters
    //tcp since we are using a networksocket instead of a Websocket
    //"localhost" is keyword for current device IP
    connection, error := net.Dial("tcp", "localhost:1")
    
    //Checks if no server available or no server matching perameters
    if error != nil {
        fmt.Println(error) //prints error if no connection
    }
    client := &Client{socket: connection}
    go client.receive() //Creates a go routine.....
    for {
        reader := bufio.NewReader(os.Stdin)
        message, _ := reader.ReadString('\n')
        connection.Write([]byte(strings.TrimRight(message, "\n")))
    }
}

func main() {
    fmt.Println("Chat Room")
    
	//ServerOrClient for command line arguments
	//must use -start when running, if not automatically "server" due to else clause
	ServerOrClient:= flag.String("start", "server", "flag title- ServerOrClient")
    flag.Parse() //needed after flags are set
    	
    if strings.ToLower(*ServerOrClient) == "client" {
        startClient()
    } else {
        startServer()
    }
}