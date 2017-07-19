package main

import (
	"github.com/miekg/dns"
	"log"
)

// DNSHandler
type DNSHandler struct {
}

// NewHandler returns a new DNSHandler
func NewHandler() *DNSHandler {

	return &DNSHandler{}
}

// DoTCP starts a tcp query
func (h *DNSHandler) DoTCP(w dns.ResponseWriter, req *dns.Msg) {
	go h.do("tcp", w, req)
}

// DoUDP starts a udp query
func (h *DNSHandler) DoUDP(w dns.ResponseWriter, req *dns.Msg) {
	go h.do("udp", w, req)
}

func (h *DNSHandler) do(Net string, w dns.ResponseWriter, req *dns.Msg) {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	defer w.Close()
}