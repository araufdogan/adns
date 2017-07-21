package main

import (
	"database/sql"
	"errors"
)

func DBGetNsByName(db *sql.DB, name string) (error, Ns) {
	var ns Ns
	err := db.QueryRow("SELECT * FROM ns WHERE name = ? LIMIT 1", name).Scan(
		&ns.Id,
		&ns.Name,
		&ns.DataV4,
		&ns.DataV6,
		&ns.Ttl)
	if err != nil || ns.Id == 0 {
		return errors.New("no row found"), ns
	}

	return nil, ns
}

func DBGetNsById(db *sql.DB, id uint32) (error, Ns) {
	var ns Ns
	err := db.QueryRow("SELECT * FROM ns WHERE id = ?", id).Scan(
		&ns.Id,
		&ns.Name,
		&ns.DataV4,
		&ns.DataV6,
		&ns.Ttl)
	if err != nil || ns.Id == 0 {
		return errors.New("no row found"), ns
	}

	return nil, ns
}

func DBGetSoaByOrigin(db *sql.DB, origin string) (error, Soa) {
	var soa Soa
	err := db.QueryRow("SELECT * FROM soa WHERE origin = ? LIMIT 1", origin).Scan(
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

	if err != nil || soa.Id == 0 {
		return errors.New("no row found"), soa
	}

	return nil, soa
}

func DBGetRrByZoneName(db *sql.DB, group string, zone uint32, name string) (error, []Rr) {
	var rr_array []Rr
	rows, err := db.Query("SELECT * FROM rr WHERE `type` = ? AND zone = ? AND name = ?", group, zone, name)
	if err != nil {
		return errors.New("no row found"), rr_array
	}

	for rows.Next() {
		var rr Rr
		err = rows.Scan(
			&rr.Id,
			&rr.Zone,
			&rr.Name,
			&rr.Data,
			&rr.Aux,
			&rr.Ttl,
			&rr.Type,
			&rr.Active,
		)

		if err != nil {
			continue
		}

		rr_array = append(rr_array, rr)
	}
	
	rows.Close()

	return nil, rr_array
}