package main

import (
	"github.com/miekg/dns"
	"log"
	"database/sql"
	"strings"
	"net"
)

// Question type
type Question struct {
	Qname  string `json:"name"`
	Qtype  string `json:"type"`
	Qclass string `json:"class"`
}

// UnFqdn function
func UnFqdn(s string) string {
	if dns.IsFqdn(s) {
		return s[:len(s)-1]
	}
	return s
}

// String formats a question
func (q *Question) String() string {
	return q.Qname + " " + q.Qclass + " " + q.Qtype
}

// DNSHandler
type DNSHandler struct {
	db *sql.DB
}

// NewHandler returns a new DNSHandler
func NewHandler(db *sql.DB) *DNSHandler {
	return &DNSHandler{db: db}
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

	q := req.Question[0]
	Q := Question{strings.ToLower(UnFqdn(q.Name)) /* convert to lower case */, dns.TypeToString[q.Qtype], dns.ClassToString[q.Qclass]}

	// print log if loglevel higher than 0
	if config.LogLevel > 0 {
		log.Printf("%s\n", Q.String())
	}

	// exit if query is not inet
	if q.Qclass != dns.ClassINET {
		return
	}

	if q.Qtype == dns.TypeA {
		// if name is ns
		var ns Ns
		err := h.db.QueryRow("SELECT * FROM ns WHERE name = ? LIMIT 1", Q.Qname).Scan(&ns.Id, &ns.Name, &ns.DataV4, &ns.DataV6, &ns.Ttl)
		if err == nil && ns.Id > 0 {
			m := new(dns.Msg)
			m.SetReply(req)

			rr_header := dns.RR_Header{Name: q.Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: ns.Ttl}
			a := &dns.A{rr_header, net.ParseIP(ns.DataV4)}
			m.Answer = append(m.Answer, a)

			w.WriteMsg(m)
			return
		}

		// if name is record

	} else if q.Qtype == dns.TypeSOA {

	}
}