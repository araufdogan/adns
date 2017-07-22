package main

import (
	"github.com/miekg/dns"
	"log"
	"database/sql"
	"strings"
	"net"
	"github.com/weppos/publicsuffix-go/publicsuffix"
	"strconv"
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
	go h.Do("tcp", w, req)
}

// DoUDP starts a udp query
func (h *DNSHandler) DoUDP(w dns.ResponseWriter, req *dns.Msg) {
	go h.Do("udp", w, req)
}

func (h *DNSHandler) Do(Net string, w dns.ResponseWriter, req *dns.Msg) {
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

	switch q.Qtype {
	case dns.TypeA:
		HandleA(h, Q, q, w, req)
		break
	case dns.TypeAAAA:
		HandleAAAA(h, Q, q, w, req)
		break
	case dns.TypeNS:
		HandleNs(h, Q, q, w, req)
		break
	case dns.TypeSOA:
		HandleSoa(h, Q, q, w, req)
		break
	case dns.TypeCNAME:
		HandleCname(h, Q, q, w, req)
		break
	case dns.TypeMX:
		HandleMx(h, Q, q, w, req)
		break
	case dns.TypeSRV:
		HandleSrv(h, Q, q, w, req)
		break
	default:
		break
	}
}

func HandleA(h *DNSHandler, Q Question, q dns.Question, w dns.ResponseWriter, req *dns.Msg) {
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
	err, soa := DBGetSoaByOrigin(h.db, domain_name.SLD + "." + domain_name.TLD)
	if err != nil || soa.Active == 0 {
		return
	}

	// find a record in rr
	err, rr_array := DBGetRrByZoneName(h.db, "A", soa.Id, domain_name.TRD)
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
}

func HandleAAAA(h *DNSHandler, Q Question, q dns.Question, w dns.ResponseWriter, req *dns.Msg) {
	// get aaaa record
	err, ns := DBGetNsByName(h.db, Q.Qname)
	if err == nil {
		// build the reply
		m := new(dns.Msg)
		m.SetReply(req)

		rr_header := dns.RR_Header{Name: q.Name, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: ns.Ttl}
		aaaa := &dns.AAAA{Hdr: rr_header, AAAA: net.ParseIP(ns.DataV6)}

		m.Answer = append(m.Answer, aaaa)

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
	err, soa := DBGetSoaByOrigin(h.db, domain_name.SLD + "." + domain_name.TLD)
	if err != nil || soa.Active == 0 {
		return
	}

	// find a record in rr
	err, rr_array := DBGetRrByZoneName(h.db, "AAAA", soa.Id, domain_name.TRD)
	if err != nil || len(rr_array) == 0 {
		return
	}

	// build the reply
	m := new(dns.Msg)
	m.SetReply(req)

	for _, rr := range rr_array {
		rr_header := dns.RR_Header{Name: q.Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: rr.Ttl}
		aaaa := &dns.AAAA{Hdr: rr_header, AAAA: net.ParseIP(rr.Data)}

		m.Answer = append(m.Answer, aaaa)
	}

	// write the reply
	w.WriteMsg(m)
	return
}

func HandleNs(h *DNSHandler, Q Question, q dns.Question, w dns.ResponseWriter, req *dns.Msg) {
	// get ns record

	// find domain name record in soa
	err, soa := DBGetSoaByOrigin(h.db, Q.Qname)
	if err != nil || soa.Active == 0 {
		return
	}

	err, ns1 := DBGetNsById(h.db, soa.Ns1)
	if err != nil {
		return
	}

	// build the reply
	m := new(dns.Msg)
	m.SetReply(req)

	rr_header := dns.RR_Header{Name: q.Name, Rrtype: dns.TypeNS, Class: dns.ClassINET, Ttl: soa.Ttl}
	ns := &dns.NS{Hdr: rr_header, Ns: ns1.Name + "."}

	m.Answer = append(m.Answer, ns)

	err, ns2 := DBGetNsById(h.db, soa.Ns2)
	if err == nil {
		rr_header := dns.RR_Header{Name: q.Name, Rrtype: dns.TypeNS, Class: dns.ClassINET, Ttl: soa.Ttl}
		ns := &dns.NS{Hdr: rr_header, Ns: ns2.Name + "."}

		m.Answer = append(m.Answer, ns)
	}

	// write the reply
	w.WriteMsg(m)
	return
}

func HandleSoa(h *DNSHandler, Q Question, q dns.Question, w dns.ResponseWriter, req *dns.Msg) {
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

func HandleCname(h *DNSHandler, Q Question, q dns.Question, w dns.ResponseWriter, req *dns.Msg) {
	// parse publicsuffix

	domain_name, err := publicsuffix.Parse(Q.Qname)
	if err != nil {
		return
	}

	// find sld + tld record in soa
	err, soa := DBGetSoaByOrigin(h.db, domain_name.SLD + "." + domain_name.TLD)
	if err != nil || soa.Active == 0 {
		return
	}

	// find cname record in rr
	err, rr_array := DBGetRrByZoneName(h.db, "CNAME", soa.Id, domain_name.TRD)
	if err != nil || len(rr_array) == 0 {
		return
	}

	// build the reply
	m := new(dns.Msg)
	m.SetReply(req)

	for _, rr := range rr_array {
		rr_header := dns.RR_Header{Name: q.Name, Rrtype: dns.TypeCNAME, Class: dns.ClassINET, Ttl: rr.Ttl}
		cname := &dns.CNAME{Hdr: rr_header, Target: rr.Data + "."}

		m.Answer = append(m.Answer, cname)
	}

	// write the reply
	w.WriteMsg(m)
	return
}

func HandleMx(h *DNSHandler, Q Question, q dns.Question, w dns.ResponseWriter, req *dns.Msg) {
	// parse publicsuffix

	domain_name, err := publicsuffix.Parse(Q.Qname)
	if err != nil {
		return
	}

	// find sld + tld record in soa
	err, soa := DBGetSoaByOrigin(h.db, domain_name.SLD + "." + domain_name.TLD)
	if err != nil || soa.Active == 0 {
		return
	}

	// find mx record in rr
	err, rr_array := DBGetRrByZoneName(h.db, "MX", soa.Id, domain_name.TRD)
	if err != nil || len(rr_array) == 0 {
		return
	}

	// build the reply
	m := new(dns.Msg)
	m.SetReply(req)

	for _, rr := range rr_array {
		split_data := strings.Split(rr.Data, " ")
		if len(split_data) != 2 {
			continue
		}

		priority, err := strconv.ParseInt(split_data[0], 10, 64)
		if err != nil {
			continue
		}

		rr_header := dns.RR_Header{Name: q.Name, Rrtype: dns.TypeMX, Class: dns.ClassINET, Ttl: rr.Ttl}
		mx := &dns.MX{Hdr: rr_header, Mx: split_data[1] + ".", Preference: uint16(priority)}

		m.Answer = append(m.Answer, mx)
	}

	// write the reply
	w.WriteMsg(m)
	return
}

func HandleSrv(h *DNSHandler, Q Question, q dns.Question, w dns.ResponseWriter, req *dns.Msg) {
	// parse publicsuffix

	domain_name, err := publicsuffix.Parse(Q.Qname)
	if err != nil {
		return
	}

	// find sld + tld record in soa
	err, soa := DBGetSoaByOrigin(h.db, domain_name.SLD + "." + domain_name.TLD)
	if err != nil || soa.Active == 0 {
		return
	}

	// find mx record in rr
	err, rr_array := DBGetRrByZoneName(h.db, "SRV", soa.Id, domain_name.TRD)
	if err != nil || len(rr_array) == 0 {
		return
	}
	/*
	// build the reply
	m := new(dns.Msg)
	m.SetReply(req)

	for _, rr := range rr_array {
		rr_header := dns.RR_Header{Name: q.Name, Rrtype: dns.TypeSRV, Class: dns.ClassINET, Ttl: rr.Ttl}
		srv := &dns.SRV{Hdr: rr_header, Targ}

		m.Answer = append(m.Answer, srv)
	}

	// write the reply
	w.WriteMsg(m)
	*/
	return
}