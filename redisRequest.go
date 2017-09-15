package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"
)

type RedisRequestType int

const (
	REQUEST_UNSUPPORTED RedisRequestType = iota
	REQUEST_REDIS_GET
	REQUEST_REDIS_MGET
	REQUEST_REDIS_SET
	REQUEST_REDIS_COMMAND
	REQUEST_REDIS_INFO
	REQUEST_REDIS_PING
)

type requestProperties struct {
	name string
}

var RequestTypeDesc = [...]requestProperties{
	REQUEST_UNSUPPORTED:   {name: "REQUEST_UNKNOWN"},
	REQUEST_REDIS_GET:     {name: "GET"},
	REQUEST_REDIS_MGET:     {name: "MGET"},
	REQUEST_REDIS_SET:     {name: "SET"},
	REQUEST_REDIS_COMMAND: {name: "COMMAND"},
	REQUEST_REDIS_INFO:    {name: "INFO"},
	REQUEST_REDIS_PING:    {name: "PING"},
}

// Helper to map a protocol string to its internal request type
type requestStringMapper struct {
	m map[string]RedisRequestType
}

func newRequestStringMapper() requestStringMapper {
	return requestStringMapper{m: make(map[string]RedisRequestType)}
}

func (m *requestStringMapper) add(name string, id RedisRequestType) {
	m.m[strings.ToUpper(name)] = id
	return
}

func (m *requestStringMapper) get(request string) RedisRequestType {
	t, ok := m.m[strings.ToUpper(request)]
	if ok != true {
		t = REQUEST_UNSUPPORTED
	}
	return t
}

var gRM requestStringMapper = newRequestStringMapper()

func init() {
	for i, v := range RequestTypeDesc {
		log.Println("Adding ", v, RedisRequestType(i))
		gRM.add(v.name, RedisRequestType(i))
	}
}

func GetRequestTypeFromString(r string) RedisRequestType {
	return gRM.get(r)
}

type RedisRequest struct {
	Name        string
	requestType RedisRequestType
	Args        [][]byte
}

func readArgument(r *bufio.Reader) ([]byte, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}

	var length int
	if _, err = fmt.Sscanf(line, "$%d\r\n", &length); err != nil {
		return nil, fmt.Errorf("invalid length for argument in %s", line)
	}

	// we know the length of the argument. Just read it.
	data, err := ioutil.ReadAll(io.LimitReader(r, int64(length)))
	if err != nil {
		return nil, err
	}
	if len(data) != length {
		return nil, fmt.Errorf("Expected length %d, received %d : data'%s'", length, len(data), data)
	}

	// Now check for trailing CR
	if b, err := r.ReadByte(); err != nil || b != '\r' {
		return nil, fmt.Errorf("Expected \\r, %s", err.Error())
	}

	// And LF
	if b, err := r.ReadByte(); err != nil || b != '\n' {
		return nil, fmt.Errorf("Expected \\n, %s", err.Error())
	}

	return data, nil
}

func parseRequest(r *bufio.Reader) (*RedisRequest, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}
	if len(line) == 0 {
		return nil, fmt.Errorf("Empty line")
	}

	var argsCount int

	if _, err := fmt.Sscanf(line, "*%d\r\n", &argsCount); err != nil {
		return nil, fmt.Errorf("invalid number of arguments in %s", line)
	}
	// All next lines are pairs of:
	//$<argument length> CR LF
	//<argument data> CR LF
	// first argument is a command name, so just convert

	firstArg, err := readArgument(r)
	if err != nil {
		return nil, err
	}

	args := make([][]byte, argsCount-1)
	for i := 0; i < argsCount-1; i += 1 {
		if args[i], err = readArgument(r); err != nil {
			return nil, err
		}
	}
	var requestType RedisRequestType = GetRequestTypeFromString(string(firstArg))
	if requestType == REQUEST_UNSUPPORTED {
		return nil, fmt.Errorf("Invalid or unsupported request")
	}

	req := &RedisRequest{
		requestType: requestType,
		Name:        strings.ToUpper(string(firstArg)),
		Args:        args,
	}
	return req, nil
}
