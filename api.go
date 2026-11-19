package main

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

//go:embed web
var fileWeb embed.FS

type server struct {
	mu sync.Mutex
}

type bukuInput struct {
	Judul           string `json:"judul"`
	TahunTerbit     int    `json:"tahun_terbit"`
	JumlahEksemplar int    `json:"jumlah_eksemplar"`
	JumlahTersedia  *int   `json:"jumlah_tersedia"`
}

type bukuOutput struct {
	ID              int    `json:"id"`
	Judul           string `json:"judul"`
	TahunTerbit     int    `json:"tahun_terbit"`
	JumlahEksemplar int    `json:"jumlah_eksemplar"`
	JumlahTersedia  int    `json:"jumlah_tersedia"`
}

type anggotaInput struct {
	Nama string `json:"nama"`
}

type anggotaOutput struct {
	ID               int       `json:"id"`
	Nama             string    `json:"nama"`
	TanggalBergabung time.Time `json:"tanggal_bergabung"`
}

type pinjamInput struct {
	BukuID    int `json:"buku_id"`
	AnggotaID int `json:"anggota_id"`
}

type pinjamanOutput struct {
	ID             int        `json:"id"`
	BukuID         int        `json:"buku_id"`
	JudulBuku      string     `json:"judul_buku"`
	AnggotaID      int        `json:"anggota_id"`
	NamaAnggota    string     `json:"nama_anggota"`
	TanggalPinjam  time.Time  `json:"tanggal_pinjam"`
	Tenggat        time.Time  `json:"tenggat"`
	TanggalKembali *time.Time `json:"tanggal_kembali"`
	Denda          int        `json:"denda"`
	Aktif          bool       `json:"aktif"`
}

func jalankanServer(alamat string) error {
	s := &server{}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/books", s.daftarBuku)
	mux.HandleFunc("GET /api/books/{id}", s.lihatBuku)
	mux.HandleFunc("POST /api/books", s.tambahBuku)
	mux.HandleFunc("DELETE /api/books/{id}", s.hapusBuku)
	mux.HandleFunc("POST /api/books/{id}/stock", s.tambahStokBuku)

	mux.HandleFunc("GET /api/members", s.daftarAnggota)
	mux.HandleFunc("GET /api/members/{id}", s.lihatAnggota)
	mux.HandleFunc("POST /api/members", s.tambahAnggota)
	mux.HandleFunc("DELETE /api/members/{id}", s.hapusAnggota)

	mux.HandleFunc("GET /api/loans", s.daftarPeminjaman)
	mux.HandleFunc("POST /api/loans", s.tambahPeminjaman)
	mux.HandleFunc("POST /api/loans/{id}/return", s.kembalikanPeminjaman)

	webFS, err := fs.Sub(fileWeb, "web")
	if err != nil {
		return err
	}
	mux.Handle("/", http.FileServer(http.FS(webFS)))

	srv := &http.Server{
		Addr:              alamat,
		Handler:           catatPermintaan(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	fmt.Printf("SERVER BERJALAN DI http://localhost%s\n", alamat)
	return srv.ListenAndServe()
}

func catatPermintaan(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mulai := time.Now()
		next.ServeHTTP(w, r)
		fmt.Printf("%s %s (%s)\n", r.Method, r.URL.RequestURI(), time.Since(mulai).Round(time.Millisecond))
	})
}

func tulisJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func tulisError(w http.ResponseWriter, status int, pesan string) {
	tulisJSON(w, status, map[string]string{"error": pesan})
}

func bacaBody(r *http.Request, data interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(data); err != nil {
		return errors.New("format JSON tidak valid")
	}
	return nil
}

func idDariPath(r *http.Request) (int, error) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		return 0, errors.New("id tidak valid")
	}
	return id, nil
}

func pesanError(err error) string {
	return strings.ToLower(err.Error())
}

func ubahBuku(b buku) bukuOutput {
	return bukuOutput{
		ID:              b.ID,
		Judul:           b.Judul,
		TahunTerbit:     b.TahunTerbit,
		JumlahEksemplar: b.Jumlaheksamplar,
		JumlahTersedia:  b.JumlahTersedia,
	}
}

func ubahAnggota(a anggota) anggotaOutput {
	return anggotaOutput{
		ID:               a.IDAnggota,
		Nama:             a.Nama,
		TanggalBergabung: a.TanggalBergabung,
	}
}

func ubahPeminjaman(p peminjaman, judulBuku, namaAnggota string) pinjamanOutput {
	hasil := pinjamanOutput{
		ID:            p.IDpinjam,
		BukuID:        p.IDBuku,
		JudulBuku:     judulBuku,
		AnggotaID:     p.IDPeminjam,
		NamaAnggota:   namaAnggota,
		TanggalPinjam: p.Tanggalpinjam,
		Tenggat:       p.Tenggat,
		Denda:         p.Denda,
		Aktif:         p.TanggalKembali.IsZero(),
	}
	if !p.TanggalKembali.IsZero() {
		kembali := p.TanggalKembali
		hasil.TanggalKembali = &kembali
	}
	return hasil
}

func judulBuku(l *library) map[int]string {
	hasil := make(map[int]string)
	for _, b := range l.Buku {
		hasil[b.ID] = b.Judul
	}
	return hasil
}

func namaAnggota(l *library) map[int]string {
	hasil := make(map[int]string)
	for _, a := range l.Anggota {
		hasil[a.IDAnggota] = a.Nama
	}
	return hasil
}

func (s *server) daftarBuku(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	l, err := muatData()
	if err != nil {
		tulisError(w, http.StatusInternalServerError, "gagal membaca data buku")
		return
	}

	cari := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("search")))
	hasil := make([]bukuOutput, 0)
	for _, b := range l.Buku {
		if cari != "" && !strings.Contains(strings.ToLower(b.Judul), cari) {
			continue
		}
		hasil = append(hasil, ubahBuku(b))
	}

	tulisJSON(w, http.StatusOK, hasil)
}

func (s *server) lihatBuku(w http.ResponseWriter, r *http.Request) {
	id, err := idDariPath(r)
	if err != nil {
		tulisError(w, http.StatusBadRequest, err.Error())
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	l, err := muatData()
	if err != nil {
		tulisError(w, http.StatusInternalServerError, "gagal membaca data buku")
		return
	}

	b, err := l.searchbukubyid(id)
	if err != nil {
		tulisError(w, http.StatusNotFound, "buku tidak ditemukan")
		return
	}

	tulisJSON(w, http.StatusOK, ubahBuku(*b))
}

func (s *server) tambahBuku(w http.ResponseWriter, r *http.Request) {
	var input bukuInput
	if err := bacaBody(r, &input); err != nil {
		tulisError(w, http.StatusBadRequest, err.Error())
		return
	}

	input.Judul = strings.TrimSpace(input.Judul)
	if input.Judul == "" {
		tulisError(w, http.StatusBadRequest, "judul tidak boleh kosong")
		return
	}
	if input.TahunTerbit <= 0 {
		tulisError(w, http.StatusBadRequest, "tahun terbit harus lebih dari 0")
		return
	}
	if input.JumlahEksemplar <= 0 {
		tulisError(w, http.StatusBadRequest, "jumlah eksemplar harus lebih dari 0")
		return
	}

	tersedia := input.JumlahEksemplar
	if input.JumlahTersedia != nil {
		tersedia = *input.JumlahTersedia
	}
	if tersedia < 0 || tersedia > input.JumlahEksemplar {
		tulisError(w, http.StatusBadRequest, "jumlah tersedia harus antara 0 dan jumlah eksemplar")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	l, err := muatData()
	if err != nil {
		tulisError(w, http.StatusInternalServerError, "gagal membaca data buku")
		return
	}

	baru := l.addbuku(buku{
		Judul:           input.Judul,
		TahunTerbit:     input.TahunTerbit,
		Jumlaheksamplar: input.JumlahEksemplar,
		JumlahTersedia:  tersedia,
	})
	if err := simpanData(l); err != nil {
		tulisError(w, http.StatusInternalServerError, "gagal menyimpan buku")
		return
	}

	tulisJSON(w, http.StatusCreated, ubahBuku(*baru))
}

func (s *server) hapusBuku(w http.ResponseWriter, r *http.Request) {
	id, err := idDariPath(r)
	if err != nil {
		tulisError(w, http.StatusBadRequest, err.Error())
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	l, err := muatData()
	if err != nil {
		tulisError(w, http.StatusInternalServerError, "gagal membaca data buku")
		return
	}

	if _, err := l.searchbukubyid(id); err != nil {
		tulisError(w, http.StatusNotFound, "buku tidak ditemukan")
		return
	}
	if err := l.removebuku(id); err != nil {
		tulisError(w, http.StatusConflict, pesanError(err))
		return
	}
	if err := simpanData(l); err != nil {
		tulisError(w, http.StatusInternalServerError, "gagal menghapus buku")
		return
	}

	tulisJSON(w, http.StatusOK, map[string]string{"message": "buku berhasil dihapus"})
}

func (s *server) tambahStokBuku(w http.ResponseWriter, r *http.Request) {
	id, err := idDariPath(r)
	if err != nil {
		tulisError(w, http.StatusBadRequest, err.Error())
		return
	}

	var input struct {
		Tambahan int `json:"tambahan"`
	}
	if err := bacaBody(r, &input); err != nil {
		tulisError(w, http.StatusBadRequest, err.Error())
		return
	}
	if input.Tambahan <= 0 {
		tulisError(w, http.StatusBadRequest, "tambahan stok harus lebih dari 0")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	l, err := muatData()
	if err != nil {
		tulisError(w, http.StatusInternalServerError, "gagal membaca data buku")
		return
	}

	if err := l.tambahjumlahbuku(id, input.Tambahan); err != nil {
		tulisError(w, http.StatusNotFound, "buku tidak ditemukan")
		return
	}
	if err := simpanData(l); err != nil {
		tulisError(w, http.StatusInternalServerError, "gagal menyimpan stok")
		return
	}

	b, _ := l.searchbukubyid(id)
	tulisJSON(w, http.StatusOK, ubahBuku(*b))
}

func (s *server) daftarAnggota(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	l, err := muatData()
	if err != nil {
		tulisError(w, http.StatusInternalServerError, "gagal membaca data anggota")
		return
	}

	hasil := make([]anggotaOutput, 0)
	for _, a := range l.Anggota {
		hasil = append(hasil, ubahAnggota(a))
	}

	tulisJSON(w, http.StatusOK, hasil)
}

func (s *server) lihatAnggota(w http.ResponseWriter, r *http.Request) {
	id, err := idDariPath(r)
	if err != nil {
		tulisError(w, http.StatusBadRequest, err.Error())
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	l, err := muatData()
	if err != nil {
		tulisError(w, http.StatusInternalServerError, "gagal membaca data anggota")
		return
	}

	a, err := l.carianggota(id)
	if err != nil {
		tulisError(w, http.StatusNotFound, "anggota tidak ditemukan")
		return
	}

	tulisJSON(w, http.StatusOK, ubahAnggota(*a))
}

func (s *server) tambahAnggota(w http.ResponseWriter, r *http.Request) {
	var input anggotaInput
	if err := bacaBody(r, &input); err != nil {
		tulisError(w, http.StatusBadRequest, err.Error())
		return
	}

	input.Nama = strings.TrimSpace(input.Nama)
	if input.Nama == "" {
		tulisError(w, http.StatusBadRequest, "nama tidak boleh kosong")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	l, err := muatData()
	if err != nil {
		tulisError(w, http.StatusInternalServerError, "gagal membaca data anggota")
		return
	}

	baru := l.tambahanggota(anggota{Nama: input.Nama})
	if err := simpanData(l); err != nil {
		tulisError(w, http.StatusInternalServerError, "gagal menyimpan anggota")
		return
	}

	tulisJSON(w, http.StatusCreated, ubahAnggota(*baru))
}

func (s *server) hapusAnggota(w http.ResponseWriter, r *http.Request) {
	id, err := idDariPath(r)
	if err != nil {
		tulisError(w, http.StatusBadRequest, err.Error())
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	l, err := muatData()
	if err != nil {
		tulisError(w, http.StatusInternalServerError, "gagal membaca data anggota")
		return
	}

	dihapus, err := l.hapusanggota(id)
	if err != nil {
		if _, errCari := l.carianggota(id); errCari != nil {
			tulisError(w, http.StatusNotFound, "anggota tidak ditemukan")
			return
		}
		tulisError(w, http.StatusConflict, pesanError(err))
		return
	}
	if err := simpanData(l); err != nil {
		tulisError(w, http.StatusInternalServerError, "gagal menghapus anggota")
		return
	}

	tulisJSON(w, http.StatusOK, map[string]string{"message": dihapus.Nama + " berhasil dihapus"})
}

func (s *server) daftarPeminjaman(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	l, err := muatData()
	if err != nil {
		tulisError(w, http.StatusInternalServerError, "gagal membaca data peminjaman")
		return
	}

	aktif := r.URL.Query().Get("active")
	hanyaAktif := aktif == "1" || aktif == "true"
	judul := judulBuku(l)
	nama := namaAnggota(l)

	hasil := make([]pinjamanOutput, 0)
	for i := len(l.Peminjaman) - 1; i >= 0; i-- {
		p := l.Peminjaman[i]
		if hanyaAktif && !p.TanggalKembali.IsZero() {
			continue
		}
		hasil = append(hasil, ubahPeminjaman(p, judul[p.IDBuku], nama[p.IDPeminjam]))
	}

	tulisJSON(w, http.StatusOK, hasil)
}

func (s *server) tambahPeminjaman(w http.ResponseWriter, r *http.Request) {
	var input pinjamInput
	if err := bacaBody(r, &input); err != nil {
		tulisError(w, http.StatusBadRequest, err.Error())
		return
	}
	if input.BukuID <= 0 || input.AnggotaID <= 0 {
		tulisError(w, http.StatusBadRequest, "buku_id dan anggota_id harus diisi")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	l, err := muatData()
	if err != nil {
		tulisError(w, http.StatusInternalServerError, "gagal membaca data peminjaman")
		return
	}

	p, err := l.tambahpeminjaman(input.BukuID, input.AnggotaID, time.Now())
	if err != nil {
		tulisError(w, http.StatusBadRequest, pesanError(err))
		return
	}
	if err := simpanData(l); err != nil {
		tulisError(w, http.StatusInternalServerError, "gagal menyimpan peminjaman")
		return
	}

	judul := judulBuku(l)
	nama := namaAnggota(l)
	tulisJSON(w, http.StatusCreated, ubahPeminjaman(p, judul[p.IDBuku], nama[p.IDPeminjam]))
}

func (s *server) kembalikanPeminjaman(w http.ResponseWriter, r *http.Request) {
	id, err := idDariPath(r)
	if err != nil {
		tulisError(w, http.StatusBadRequest, err.Error())
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	l, err := muatData()
	if err != nil {
		tulisError(w, http.StatusInternalServerError, "gagal membaca data peminjaman")
		return
	}

	p, err := l.caripinjaman(id)
	if err != nil {
		tulisError(w, http.StatusNotFound, "peminjaman tidak ditemukan")
		return
	}
	if !p.TanggalKembali.IsZero() {
		tulisError(w, http.StatusConflict, "buku sudah dikembalikan")
		return
	}

	hasil, err := l.kembalikanbuku(id, time.Now())
	if err != nil {
		tulisError(w, http.StatusConflict, pesanError(err))
		return
	}
	if err := simpanData(l); err != nil {
		tulisError(w, http.StatusInternalServerError, "gagal menyimpan pengembalian")
		return
	}

	judul := judulBuku(l)
	nama := namaAnggota(l)
	tulisJSON(w, http.StatusOK, ubahPeminjaman(hasil, judul[hasil.IDBuku], nama[hasil.IDPeminjam]))
}
