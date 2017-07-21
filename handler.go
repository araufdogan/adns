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
		err := h.db.QueryRow("SELECT * FROM ns WHERE name = ? LIMIT 1", Q.Qname).Scan(
			&ns.Id,
			&ns.Name,
			&ns.DataV4,
			&ns.DataV6,
			&ns.Ttl)

		if err == nil && ns.Id > 0 {
			// build the reply
			m := new(dns.Msg)
			m.SetReply(req)

			rr_header := dns.RR_Header{Name: q.Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: ns.Ttl}
			a := &dns.A{Hdr: rr_header, A: net.ParseIP(ns.DataV4)}

			m.Answer = append(m.Answer, a)

			// write the reply
			w.WriteMsg(m)
			return
		}

		// todo if name is record

	} else if q.Qtype == dns.TypeSOA {
		// get soa record
		var soa Soa
		err := h.db.QueryRow("SELECT * FROM soa WHERE origin = ? LIMIT 1", Q.Qname).Scan(
			&soa.Id,
			&soa.UserId,
			&soa.Origin,
			&soa.Ns1,
			&soa.Ns2,
			&soa.Mbox,
			&soa.Serial,
			&soa.Refresh,
			&soa.Retry,
			&soa.Expire,
			&soa.Minimum,
			&soa.Ttl,
			&soa.Active,
		)

		if err == nil && soa.Id > 0 {
			// get ns records
			var ns1 Ns
			err := h.db.QueryRow("SELECT * FROM ns WHERE id = ?", soa.Ns1).Scan(
				&ns1.Id,
				&ns1.Name,
				&ns1.DataV4,
				&ns1.DataV6,
				&ns1.Ttl)
			if err != nil || ns1.Id == 0 {
				return
			}

			var ns2 Ns
			err = h.db.QueryRow("SELECT * FROM ns WHERE id = ?", soa.Ns2).Scan(
				&ns2.Id,
				&ns2.Name,
				&ns2.DataV4,
				&ns2.DataV6,
				&ns2.Ttl)
			if err != nil || ns2.Id == 0 {
				return
			}

			// build the reply
			m:= new(dns.Msg)
			m.SetReply(req)

			rr_header := dns.RR_Header{ Name: q.Name, Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: soa.Ttl }
			s := &dns.SOA{ Hdr: rr_header, Ns: ns1.Name + ".", Mbox: soa.Mbox + ".", Serial: soa.Serial, Refresh: soa.Refresh, Retry: soa.Retry, Expire: soa.Expire, Minttl: soa.Minimum}

			m.Answer = append(m.Answer, s)

			// write the reply
			w.WriteMsg(m)
			return
		}
	}
}