package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
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

//listen on BcChannel, and send to all agents once there is message on it
func (this *Server) listenBcChannel() {
	for {
		msg := <-this.bcChannel

		this.mapLock.Lock()
		for _, agent := range this.onlineMap {
			agent.channel <- msg
		}
		this.mapLock.Unlock()
	}
}

func (this *Server) broadcast(agent *Agent, msg string) {
	sendMsg := "[" + agent.userAddr + "]" + agent.userName + ": " + msg
	this.bcChannel <- sendMsg
}

func (this *Server) updateUserName(agent *Agent, newName string) {
	this.mapLock.Lock()
	_, ok := this.onlineMap[newName]
	this.mapLock.Unlock()
	if ok {
		agent.sendMsg("The name is already used\n")
	} else {
		fmt.Println("[" + agent.userAddr + "]" + agent.userName + " updated user name to: " + newName)

		this.mapLock.Lock()
		delete(this.onlineMap, agent.userName)
		this.onlineMap[newName] = agent
		this.mapLock.Unlock()

		agent.userName = newName
		agent.sendMsg("user name updated to " + agent.userName + "\n")
	}
}

func (this *Server) handler(conn net.Conn) {
	//add user - agent to map
	agent := newAgent(conn, this)

	agent.online()

	isLive := make(chan bool)

	// accept client's message
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				agent.offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("conn.Read err: ", err)
				return
			}

			msg := string(buf[:n-1])
			agent.handleMessage(msg)

			//any message infers user is live
			isLive <- true
		}
	}()

	//block handler
	for {
		select {
		case <-isLive:
			//the current user is live
		case <-time.After(time.Second * 20):
			//time expired. kick out user
			agent.sendMsg("You are kicked. Press ENTER to exit\n")
			//agent.offline()
			close(agent.channel)
			conn.Close()
			return
		}
	}
}

func (this *Server) start() {
	//listen (close listener)
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.ip, this.port))
	if err != nil {
		fmt.Println("net.Listen error: ", err)
	}
	defer listener.Close()

	go this.listenBcChannel()

	for {
		//accpet
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener.accept error: ", err)
			continue
		}
		//handler
		go this.handler(conn)
	}
	//close

}

func main() {
	server := newServer("127.0.0.1", 8888)
	server.start()
}
