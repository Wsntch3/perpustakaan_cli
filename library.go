package main

import (
	"errors"
	"strings"
	"time"
)

type library struct {
	Buku            []buku       `json:"Buku"`
	Anggota         []anggota    `json:"Anggota"`
	Peminjaman      []peminjaman `json:"Peminjaman"`
	NewIDpeminjaman int          `json:"NewIDpeminjaman"`
	NewIDanggota    int          `json:"NewIdAnggota"`
	NewId           int          `json:"NewId"`
}

func Defaultscore() *library {
	return &library{
		Buku:            []buku{},
		Anggota:         []anggota{},
		Peminjaman:      []peminjaman{},
		NewIDpeminjaman: 1,
		NewIDanggota:    1,
		NewId:           1,
	}
}

func (l *library) addbuku(b buku) *buku {
	b.ID = l.NewId
	l.NewId++
	l.Buku = append(l.Buku, b)
	return &l.Buku[len(l.Buku)-1]
}

func (l *library) removebuku(id int) error {
	for i := range l.Buku {
		if l.Buku[i].ID == id {
			l.Buku = append(l.Buku[:i], l.Buku[i+1:]...)
			return nil
		}
	}
	return errors.New("BUKU GAGAL DIHAPUS")
}

func (l *library) searchbuku(judulbuku string) (*buku, error) {
	for i := range l.Buku {
		var namakecil string = strings.ToLower(l.Buku[i].Judul)
		var searchkecil string = strings.ToLower(judulbuku)

		if strings.Contains(namakecil, searchkecil) {
			return &l.Buku[i], nil
		}
	}
	return nil, errors.New("BUKU TIDAK DITEMUKAN")
}

func (l *library) searchbukubyid(id int) (*buku, error) {
	for i := range l.Buku {
		if l.Buku[i].ID == id {
			return &l.Buku[i], nil
		}
	}
	return nil, errors.New("BUKU TIDAK DITEMUKAN")
}

func (l *library) lihatsemuabuku() []buku {

	return l.Buku
}

func (l *library) tambahanggota(a anggota) *anggota {
	a.IDAnggota = l.NewIDanggota
	l.NewIDanggota++
	a.TanggalBergabung = time.Now()
	l.Anggota = append(l.Anggota, a)
	return &l.Anggota[len(l.Anggota)-1]
}

func (l *library) tambahjumlahbuku(id, tambahan int) error {
	if tambahan <= 0 {
		return errors.New("TAMBAHAN STOK HARUS LEBIH DARI 0")
	}

	for i := range l.Buku {
		if l.Buku[i].ID == id {
			l.Buku[i].Jumlaheksamplar += tambahan
			l.Buku[i].JumlahTersedia += tambahan
			return nil
		}
	}
	return errors.New("STOK BUKU GAGAL DITAMBAHKAN")
}

func (l *library) carianggota(id int) (*anggota, error) {
	for i := range l.Anggota {
		if l.Anggota[i].IDAnggota == id {
			return &l.Anggota[i], nil
		}
	}
	return nil, errors.New("ANGGOTA TIDAK DITEMUKAN")
}

func (l *library) tampilkananggota() []anggota {
	return l.Anggota
}

func (l *library) hapusanggota(id int) (*anggota, error) {
	for i := range l.Anggota {
		if l.Anggota[i].IDAnggota == id {
			dihapus := l.Anggota[i]
			l.Anggota = append(l.Anggota[:i], l.Anggota[i+1:]...)
			return &dihapus, nil
		}
	}
	return nil, errors.New("ANGGOTA TIDAK DITEMUKAN")
}

func (l *library) hitungaktif(idAnggota int) int {
	var hitungtotal int = 0
	for i := range l.Peminjaman {
		if l.Peminjaman[i].IDPeminjam == idAnggota && l.Peminjaman[i].TanggalKembali.IsZero() {
			hitungtotal++

		}
	}
	return hitungtotal
}

func (l *library) adaterlambat(idAnggota int, sekarang time.Time) bool {
	for i := range l.Peminjaman {
		if l.Peminjaman[i].IDPeminjam == idAnggota {
			if sekarang.After(l.Peminjaman[i].Tenggat) && l.Peminjaman[i].TanggalKembali.IsZero() {
				return true
			}
		}
	}
	return false
}

func (l *library) caripinjaman(idpinjaman int) (*peminjaman, error) {
	for i := range l.Peminjaman {
		if l.Peminjaman[i].IDpinjam == idpinjaman {
			return &l.Peminjaman[i], nil
		}
	}
	return nil, errors.New("ID PEMINJAMAN TIDAK TERSEDIA")
}
func (l *library) tambahpeminjaman(id int, idanggota int, sekarang time.Time) (peminjaman, error) {
	cekanggota, err := l.carianggota(idanggota)
	if err != nil {
		return peminjaman{}, err
	}
	cekbuku, err := l.searchbukubyid(id)
	if err != nil {
		return peminjaman{}, err
	}
	if cekbuku.JumlahTersedia < 1 {
		return peminjaman{}, errors.New("BUKU TIDAK TERSEDIA")
	}
	cekpinjamanaktif := l.hitungaktif(idanggota)
	if cekpinjamanaktif >= 3 {
		return peminjaman{}, errors.New("BATAS 3 BUKU")
	}
	cekketerlambatan := l.adaterlambat(idanggota, sekarang)
	if cekketerlambatan {
		return peminjaman{}, errors.New("ADA BUKU TERLAMBAT")
	}

	cekbuku.JumlahTersedia = cekbuku.JumlahTersedia - 1

	p := peminjaman{
		IDpinjam:      l.NewIDpeminjaman,
		IDPeminjam:    cekanggota.IDAnggota,
		IDBuku:        cekbuku.ID,
		Tanggalpinjam: sekarang,
		Tenggat:       sekarang.AddDate(0, 0, 3),
	}
	l.NewIDpeminjaman++
	l.Peminjaman = append(l.Peminjaman, p)

	return p, nil
}

func denda(tenggat, tanggalkembali time.Time) int {
	if !tanggalkembali.After(tenggat) {
		return 0
	}

	selisih := tanggalkembali.Sub(tenggat)
	hari := int(selisih.Hours() / 24)

	return hari * 1000
}

func (l *library) kembalikanbuku(idpinjam int, tanggalkembali time.Time) (peminjaman, error) {
	pinjam, err := l.caripinjaman(idpinjam)
	if err != nil {
		return peminjaman{}, err
	}
	if !pinjam.TanggalKembali.IsZero() {
		return peminjaman{}, errors.New("BUKU SUDAH DIKEMBALIKAN")
	}
	pinjam.TanggalKembali = tanggalkembali
	pinjam.Denda = denda(pinjam.Tenggat, tanggalkembali)
	buku, errc := l.searchbukubyid(pinjam.IDBuku)
	if errc == nil && buku.JumlahTersedia < buku.Jumlaheksamplar {
		buku.JumlahTersedia++
	}
	return *pinjam, nil
}

func (l *library) daftarpeminjamanaktif() []peminjaman {
	var hasil []peminjaman

	for i := range l.Peminjaman {
		if l.Peminjaman[i].TanggalKembali.IsZero() {
			hasil = append(hasil, l.Peminjaman[i])
		}
	}
	return hasil
}
