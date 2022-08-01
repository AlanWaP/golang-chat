package main

import (
	"fmt"
	"net"
	"strings"
)

type Agent struct {
	userName string
	userAddr string
	channel  chan string
	conn     net.Conn
	server   *Server
}

func newAgent(conn net.Conn, server *Server) *Agent {
	userAddr := conn.RemoteAddr().String()
	agent := &Agent{
		userName: userAddr,
		userAddr: userAddr,
		channel:  make(chan string),
		conn:     conn,
		server:   server,
	}

	go agent.ListenMessage()

	return agent
}

func (this *Agent) online() {
	this.server.mapLock.Lock()
	this.server.onlineMap[this.userName] = this
	this.server.mapLock.Unlock()

	//broadcast new user info
	fmt.Println("[" + this.userAddr + "]" + this.userName + ": online")
	this.server.broadcast(this, "online")
}

func (this *Agent) offline() {
	this.server.mapLock.Lock()
	delete(this.server.onlineMap, this.userName)
	this.server.mapLock.Unlock()

	//broadcast new user info
	fmt.Println("[" + this.userAddr + "]" + this.userName + ": offline")
	this.server.broadcast(this, "offline")
}

func (this *Agent) sendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

func (this *Agent) handleMessage(msg string) {
	if msg == "who" {
		this.server.mapLock.Lock()
		for _, agent := range this.server.onlineMap {
			onlineMsg := "[" + agent.userAddr + "]" + agent.userName + ": online\n"
			this.sendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		newName := strings.Split(msg, "|")[1]
		//newName := msg[7:]
		this.server.updateUserName(this, newName)
	} else if msg == "me" {
		this.sendMsg("You are [" + this.userAddr + "]" + this.userName + "\n")
	} else if len(msg) > 4 && msg[:3] == "to|" && len(strings.Split(msg, "|")) >= 3 {
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			this.sendMsg("format incorrect. Should be to|<remoteName>|<message>\n")
			return
		}
		this.server.mapLock.Lock()
		remoteUser, ok := this.server.onlineMap[remoteName]
		this.server.mapLock.Unlock()
		if !ok {
			this.sendMsg("Remote Name does not exist\n")
			return
		}
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.sendMsg("Content is empty. Please resend\n")
			return
		}
		remoteUser.sendMsg("[" + this.userAddr + "]" + this.userName + ": " + content + "\n")
	} else {
		this.server.broadcast(this, msg)
	}
}

func (this *Agent) ListenMessage() {
	for {
		msg := <-this.channel
		this.conn.Write([]byte(msg + "\n"))
	}
}
