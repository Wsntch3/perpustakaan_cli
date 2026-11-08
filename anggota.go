package main

import "time"

type anggota struct {
	IDAnggota        int       `json:"IDAnggota"`
	Nama             string    `json:"Nama"`
	TanggalBergabung time.Time `json:"TanggalBergabung"`
}
