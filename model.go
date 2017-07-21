package main

type Ns struct {
	Id		uint32
	Name		string
	DataV4		string
	DataV6		string
	Ttl		uint32
}

type Soa struct {
	Id		uint32
	UserId		uint32
	Origin		string
	Ns1		uint32
	Ns2		uint32
	Mbox		string
	Serial		uint32
	Refresh		uint32
	Retry		uint32
	Expire		uint32
	Minimum		uint32
	Ttl		uint32
	Active		string
}