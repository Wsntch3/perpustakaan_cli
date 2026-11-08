package main

type buku struct {
	ID              int    `json:"ID"`
	Judul           string `json:"Judul"`
	TahunTerbit     int    `json:"TahunTerbit"`
	Jumlaheksamplar int    `json:"Jumlaheksamplar"`
	JumlahTersedia  int    `json:"JumlahTersedia"`
}
