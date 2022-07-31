package main

import (
	"fmt"
	"net"
	"sync"
)

type Server struct {
	ip   string
	port int
	//map of online user to agent and its lock
	onlineMap map[string]*Agent
	mapLock   sync.RWMutex
	//broadcast channel
	bcChannel chan string
}

func newServer(ip string, port int) *Server {
	return &Server{
		ip:        ip,
		port:      port,
		onlineMap: make(map[string]*Agent),
		bcChannel: make(chan string),
	}
}

//listen BcChannel, and send to all agents once there is message on it
func (server *Server) listenBcChannel() {
	for {
		msg := <-server.bcChannel

		server.mapLock.Lock()
		for _, agent := range server.onlineMap {
			agent.channel <- msg
		}
		server.mapLock.Unlock()
	}
}

func (server *Server) broadcast(agent *Agent, msg string) {
	sendMsg := "[" + agent.userAddr + "]" + agent.userName + ": " + msg
	server.bcChannel <- sendMsg
}

func (server *Server) handler(conn net.Conn) {
	//Handle
	fmt.Println("new connection established")

	//add user - agent to map
	agent := newAgent(conn)
	server.mapLock.Lock()
	server.onlineMap[agent.userName] = agent
	server.mapLock.Unlock()

	//broadcast new user info
	server.broadcast(agent, "online")

	//block handler
	select {}
}

func (server *Server) start() {
	//listen (close listener)
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.ip, server.port))
	if err != nil {
		fmt.Println("net.Listen error: ", err)
	}
	defer listener.Close()

	go server.listenBcChannel()

	for {
		//accpet
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener.accept error: ", err)
			continue
		}
		//handler
		go server.handler(conn)
	}
	//close

}

func main() {
	server := newServer("127.0.0.1", 8888)
	server.start()
}
