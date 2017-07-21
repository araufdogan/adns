package main

import (
	"github.com/miekg/dns"
	"log"
	"database/sql"
	"strings"
	"net"
	"github.com/weppos/publicsuffix-go/publicsuffix"
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
		err, ns := DBGetNsByName(h.db, Q.Qname)
		if err == nil {
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

		// if name is record
		// parse publicsuffix

		domain_name, err := publicsuffix.Parse(Q.Qname)
		if err != nil {
			return
		}

		// find sld + tld record in soa
		err, soa := DBGetSoaByOrigin(h.db, domain_name.SLD + domain_name.TLD)
		if err != nil || soa.Active == 0 {
			return
		}

		// find a record in rr
		err, rr_array := DBGetRrByZoneName(h.db, "A", soa.Id, domain_name.SLD)
		if err != nil || len(rr_array) == 0 {
			return
		}

		// build the reply
		m := new(dns.Msg)
		m.SetReply(req)

		for _, rr := range rr_array {
			rr_header := dns.RR_Header{Name: q.Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: rr.Ttl}
			a := &dns.A{Hdr: rr_header, A: net.ParseIP(rr.Data)}

			m.Answer = append(m.Answer, a)
		}

		// write the reply
		w.WriteMsg(m)
		return
	} else if q.Qtype == dns.TypeSOA {
		// get soa record
		err, soa := DBGetSoaByOrigin(h.db, Q.Qname)
		if err != nil || soa.Active == 0 {
			return
		}

		err, ns1 := DBGetNsById(h.db, soa.Ns1)
		if err != nil {
			return
		}

		/*
		err, ns2 := DBGetNsById(h.db, soa.Ns2)
		if err != nil {
			return
		}
		*/

		// build the reply
		m := new(dns.Msg)
		m.SetReply(req)

		rr_header := dns.RR_Header{ Name: q.Name, Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: soa.Ttl }
		s := &dns.SOA{ Hdr: rr_header, Ns: ns1.Name + ".", Mbox: soa.Mbox + ".", Serial: soa.Serial, Refresh: soa.Refresh, Retry: soa.Retry, Expire: soa.Expire, Minttl: soa.Minimum}

		m.Answer = append(m.Answer, s)

		// write the reply
		w.WriteMsg(m)
		return
	}
}