# simple chat server
Learn from [here](https://bilibili.com/video/BV1gf4y1r79E).    
Run `go build -o server server.go agent.go && ./server` to start a server    

 To start a client, you can either run `nc 127.0.0.1 8888`
- Type `me` to check address and name   
- Type `who` to check all live users   
- Type `rename|<new name>` to change user name   
- Type `to|<remote name>|<message>` to chat to single user    
- Type `<message>` to broadcast      

Or you can run `go build client.go && ./client` with better UI
- The message from server is green 