package main

import "time"

type peminjaman struct {
	IDpinjam       int       `json:"IDpinjam"`
	IDPeminjam     int       `json:"IDPeminjam"`
	IDBuku         int       `json:"IDbukudipinjam"`
	Tanggalpinjam  time.Time `json:"Tanggalpinjam"`
	Tenggat        time.Time `json:"Tenggat"`
	TanggalKembali time.Time `json:"TanggalKembali"`
	Denda          int       `json:"Denda"`
}
