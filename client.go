package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
)

type Client struct {
	serverIp   string
	serverPort int
	conn       net.Conn
	flag       int
}

func newClient(serverIp string, serverPort int) *Client {
	client := &Client{
		serverIp:   serverIp,
		serverPort: serverPort,
		flag:       -1,
	}
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error: ", err)
		return nil
	}
	client.conn = conn
	return client
}

//handle response from server. Print to stdio
func (this *Client) handleResponse() {
	for {
		buf := make([]byte, 4096)
		len, err := this.conn.Read(buf)
		if len == 0 || err != nil {
			this.flag = 0
			return
		} else {
			fmt.Print(string("\033[32m"), string(buf), string("\033[0m"))
		}
	}
	//io.Copy(os.Stdout, this.conn)
	/* The above statement copies data from this.conn to stdout and is permanently blocked. It is equal to the following:
	for {
		buf := make([]byte, 4096)
		client.conn.Read(buf)
		fmt.Println(buf)
	}
	*/
}

func (this *Client) menu() bool {
	fmt.Println("Please input a number: ")
	fmt.Println("1. Broadcast")
	fmt.Println("2. Chat")
	fmt.Println("3. Update user name")
	fmt.Println("0. Exit")
	var flag int
	_, err := fmt.Scanln(&flag)
	if this.flag == 0 {
		return true
	}
	if err != nil {
		fmt.Println("Please input a valid number!!!")
		return false
	}
	if flag >= 0 && flag <= 3 {
		this.flag = flag
		return true
	} else {
		fmt.Println("Please input a valid number!!!")
		return false
	}
}

func (this *Client) broadcast() bool {
	reader := bufio.NewReader(os.Stdin)
	// fmt.Scanln(&msg) is not good enough
	fmt.Println("Please input broadcast content. 'exit' to leave")
	buf, _ := reader.ReadBytes('\n')
	for string(buf) != "exit\n" {
		if this.flag == 0 {
			return false
		}
		_, err := this.conn.Write(buf)
		if err != nil {
			fmt.Println("conn.Write err: ", err)
			return false
		}
		fmt.Println("Please input broadcast content. 'exit' to leave")
		buf, _ = reader.ReadBytes('\n')
	}
	return true
}

func (this *Client) checkUsers() bool {
	msg := "who\n"
	_, err := this.conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("conn.Write error:", err)
		return false
	}
	return true
}

func (this *Client) chat() bool {
	if this.flag == 0 || !this.checkUsers() {
		return false
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Please select user. Input user name. 'exit' to leave")
	buf, _ := reader.ReadBytes('\n')
	remoteName := string(buf[:len(buf)-1])
	for remoteName != "exit" {
		if this.flag == 0 {
			return false
		}
		fmt.Println("Please input chat content (should not contain '|'). 'exit' to leave")
		buf, _ = reader.ReadBytes('\n')
		msg := string(buf[:len(buf)-1])
		for msg != "exit" {
			if this.flag == 0 {
				return false
			}
			if len(msg) != 0 {
				msg = "to|" + remoteName + "|" + msg + "\n"
				_, err := this.conn.Write([]byte(msg))
				if err != nil {
					fmt.Println("conn.Write error:", err)
					return false
				}
			}
			msg = ""
			fmt.Println("Please input chat content. 'exit' to leave, or when you find remote user does not exist")
			buf, _ = reader.ReadBytes('\n')
			msg = string(buf[:len(buf)-1])
		}
		if !this.checkUsers() {
			return false
		}
		fmt.Println("Please select user. Input user name. 'exit' to leave")
		buf, _ := reader.ReadBytes('\n')
		remoteName = string(buf[:len(buf)-1])
	}
	return true
}

func (this *Client) updateName() bool {
	fmt.Println("Please input your new name (should not contain space or '|'). 'exit' to leave")
	var newName string
	fmt.Scanln(&newName)
	if newName == "exit" {
		return true
	}
	msg := "rename|" + newName + "\n"
	_, err := this.conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("conn.Write error: ", err)
		return false
	}
	return true
}

func (this *Client) run() {
	for this.flag != 0 {
		for !this.menu() {
		}
		switch this.flag {
		case 1:
			fmt.Println("Broadcast...")
			if !this.broadcast() {
				return
			}
		case 2:
			fmt.Println("Chat...")
			if !this.chat() {
				return
			}
		case 3:
			fmt.Println("Update user name...")
			if !this.updateName() {
				return
			}
		case 0:
			fmt.Println("Exit...")
		}
	}
}

var serverIp string
var serverPort int

// ./client -ip 127.0.0.1 -port 8888
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "set server IP")
	flag.IntVar(&serverPort, "port", 8888, "set server port")
}

func main() {
	flag.Parse()

	client := newClient(serverIp, serverPort)
	if client == nil {
		fmt.Println("Connection failed...")
		return
	}

	fmt.Println("Connection success!")
	go client.handleResponse()
	client.run()
}
