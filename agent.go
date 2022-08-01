package main

import (
	"fmt"
	"net"
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

func (this *Agent) handleMessage(msg string) {
	this.server.broadcast(this, msg)
}

func (this *Agent) ListenMessage() {
	for {
		msg := <-this.channel
		this.conn.Write([]byte(msg + "\n"))
	}
}
