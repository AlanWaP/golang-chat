package main

import (
	"net"
)

type Agent struct {
	userName string
	userAddr string
	channel  chan string
	conn     net.Conn
}

func newAgent(conn net.Conn) *Agent {
	userAddr := conn.RemoteAddr().String()
	agent := &Agent{
		userName: userAddr,
		userAddr: userAddr,
		channel:  make(chan string),
		conn:     conn,
	}

	go agent.ListenMessage()

	return agent
}

func (this *Agent) ListenMessage() {
	for {
		msg := <-this.channel
		this.conn.Write([]byte(msg + "\n"))
	}
}
