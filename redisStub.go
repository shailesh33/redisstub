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
	log.Printf("Received %s", req.Name)
	switch req.requestType {
	case REQUEST_REDIS_COMMAND:
		rsp := newArrayResponse()
		rsp.Write(w)
	case REQUEST_REDIS_PING:
		rsp := newStringResponse([]byte{'P', 'O', 'N', 'G'})
		rsp.Write(w)
	case REQUEST_REDIS_INFO:
		rsp := newArrayResponse()
		rsp.Write(w)
	case REQUEST_REDIS_GET:
		fallthrough
	case REQUEST_REDIS_MGET:
		fallthrough
	default:
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
