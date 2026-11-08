package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func savedata(file string, l *library) error {
	data, err := json.MarshalIndent(l, "", " ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
		return err
	}

	return os.WriteFile(file, data, 0644)
}

func loaddata(file string) (*library, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			return Defaultscore(), nil
		}
		return nil, err
	}

	var l library
	if err := json.Unmarshal(data, &l); err != nil {
		return nil, err
	}

	if l.Buku == nil {
		l.Buku = []buku{}
	}
	if l.Anggota == nil {
		l.Anggota = []anggota{}
	}
	if l.Peminjaman == nil {
		l.Peminjaman = []peminjaman{}
	}

	// Pastikan ID berikutnya selalu lebih besar dari ID yang sudah ada,
	// termasuk untuk file lama yang belum memiliki counter dengan benar.
	maxIDBuku := 0
	for _, b := range l.Buku {
		if b.ID > maxIDBuku {
			maxIDBuku = b.ID
		}
	}
	if l.NewId <= maxIDBuku {
		l.NewId = maxIDBuku + 1
	}

	maxIDAnggota := 0
	for _, a := range l.Anggota {
		if a.IDAnggota > maxIDAnggota {
			maxIDAnggota = a.IDAnggota
		}
	}
	if l.NewIDanggota <= maxIDAnggota {
		l.NewIDanggota = maxIDAnggota + 1
	}

	maxIDPeminjaman := 0
	for _, p := range l.Peminjaman {
		if p.IDpinjam > maxIDPeminjaman {
			maxIDPeminjaman = p.IDpinjam
		}
	}
	if l.NewIDpeminjaman <= maxIDPeminjaman {
		l.NewIDpeminjaman = maxIDPeminjaman + 1
	}

	return &l, nil
}
