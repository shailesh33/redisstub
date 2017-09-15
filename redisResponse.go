package main

import (
	"bufio"
	"strconv"
)

type Writable interface {
	Write(w *bufio.Writer)
}

type nilResponse struct {
}

func newNilResponse() *nilResponse {
	return &nilResponse{}
}

func (r nilResponse) Write(w *bufio.Writer) error {
	w.WriteString("$-1\r\n")
	w.Flush()
	return nil
}

/////////// integer response
type integerResponse struct {
	I int
}

func newIntegerResponse(i int) *integerResponse {
	return &integerResponse{
		I: i,
	}
}

func (r integerResponse) Write(w *bufio.Writer) error {
	w.WriteByte(':')
	w.WriteString(strconv.Itoa(r.I))
	w.WriteString("\r\n")
	w.Flush()
	return nil
}

//////////// Status response
type statusResponse struct {
	S string
}

func newStatusResponse(s string) *statusResponse {
	return &statusResponse{
		S: s,
	}
}

func (r statusResponse) Write(w *bufio.Writer) error {
	w.WriteString("+" + r.S)
	w.WriteString("\r\n")
	w.Flush()
	return nil
}

/////////////// error response
// error response
type errorResponse struct {
	errorString string
}

func newErrorResponse(s string) *errorResponse {
	return &errorResponse{
		errorString: s,
	}
}

func (r errorResponse) Write(w *bufio.Writer) error {
	w.WriteString("-" + r.errorString)
	w.WriteString("\r\n")
	w.Flush()
	return nil
}

/////////////// string response
type stringResponse struct {
	data []byte
}

func newStringResponse(b []byte) *stringResponse {
	return &stringResponse{
		data: b,
	}
}

func (r stringResponse) Write(w *bufio.Writer) error {
	w.WriteString("$" + strconv.Itoa(len(r.data)) + "\r\n")
	w.Write(r.data)
	w.WriteString("\r\n")
	w.Flush()
	return nil
}

//////////////// array response
type arrayResponse struct {
	elems []Writable
}

func (r arrayResponse) Write(w *bufio.Writer) error {
	w.WriteByte('*')
	w.WriteString(strconv.Itoa(len(r.elems)))
	w.WriteString("\r\n")
	for _, i := range r.elems {
		i.Write(w)
	}
	w.Flush()
	return nil
}

func (r arrayResponse) AppendArgs(elem Writable) {
	r.elems = append(r.elems, elem)
}

func newArrayResponse() *arrayResponse {
	return &arrayResponse{}
}
