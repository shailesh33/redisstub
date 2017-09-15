package main

import (
	"bufio"
	"flag"
	"log"
	"net"
	"os"
	"strconv"
)

var port int

func init() {
	flag.IntVar(&port, "p", 22122, "Port to listen to")
	flag.Parse()
}

func respond(conn net.Conn, req *RedisRequest) {
	w := bufio.NewWriter(conn)
	switch req.requestType {
	case REQUEST_REDIS_COMMAND:
		log.Printf("received command")
		rsp := newArrayResponse()
		rsp.Write(w)
	case REQUEST_REDIS_MGET:
		log.Printf("Received MGET")
		rsp := newErrorResponse("Storage: Too many arguments")
		rsp.Write(w)
	}
}

func redisClientConnHandler(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	for {
		req, err := parseRequest(r)
		if err != nil {
			log.Printf("Error parsing %s", err.Error())
			return
		}
		log.Printf("Received %+v", req)
		respond(conn, req)
	}

}

func main() {
	listener, err := net.Listen("tcp", net.JoinHostPort("localhost", strconv.Itoa(port)))
	if err != nil {
		log.Println("Error listening on ", port, err.Error())
		os.Exit(1)
	}
	defer listener.Close()
	log.Println("Listening on ", port)
	for {
		// Listen for an incoming connection.
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go redisClientConnHandler(conn)
	}
}
