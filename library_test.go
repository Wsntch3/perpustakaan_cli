	package main

	import (
		"os"
		"path/filepath"
		"testing"
		"time"
	)

	func TestAddBukuMengembalikanBukuBaru(t *testing.T) {
		l := Defaultscore()

		b := l.addbuku(buku{Judul: "Buku A", Jumlaheksamplar: 2, JumlahTersedia: 2})

		if b == nil {
			t.Fatal("addbuku mengembalikan nil")
		}
		if b.ID != 1 || b.Judul != "Buku A" {
			t.Fatalf("buku baru tidak sesuai: %+v", b)
		}
		if len(l.Buku) != 1 {
			t.Fatalf("jumlah buku = %d, ingin 1", len(l.Buku))
		}
	}

	func TestTambahAnggotaMengembalikanAnggotaBaru(t *testing.T) {
		l := Defaultscore()

		l.tambahanggota(anggota{Nama: "Anggota A"})
		a := l.tambahanggota(anggota{Nama: "Anggota B"})

		if a.Nama != "Anggota B" {
			t.Fatalf("anggota yang dikembalikan = %q, ingin %q", a.Nama, "Anggota B")
		}
		if a.IDAnggota != 2 {
			t.Fatalf("ID anggota = %d, ingin 2", a.IDAnggota)
		}
		if a.TanggalBergabung.IsZero() {
			t.Fatal("tanggal bergabung tidak diisi")
		}
	}

	func TestHapusAnggotaTerakhir(t *testing.T) {
		l := Defaultscore()
		l.tambahanggota(anggota{Nama: "Anggota A"})
		terakhir := l.tambahanggota(anggota{Nama: "Anggota B"})

		dihapus, err := l.hapusanggota(terakhir.IDAnggota)
		if err != nil {
			t.Fatalf("hapusanggota error: %v", err)
		}
		if dihapus.Nama != "Anggota B" {
			t.Fatalf("anggota yang dihapus = %q, ingin %q", dihapus.Nama, "Anggota B")
		}
		if len(l.Anggota) != 1 || l.Anggota[0].Nama != "Anggota A" {
			t.Fatalf("sisa anggota tidak sesuai: %+v", l.Anggota)
		}
	}

	func TestHapusAnggotaTidakDitemukan(t *testing.T) {
		l := Defaultscore()
		l.tambahanggota(anggota{Nama: "Anggota A"})

		if _, err := l.hapusanggota(99); err == nil {
			t.Fatal("hapusanggota seharusnya mengembalikan error")
		}
		if len(l.Anggota) != 1 {
			t.Fatalf("jumlah anggota berubah menjadi %d", len(l.Anggota))
		}
	}

	func TestRemoveBuku(t *testing.T) {
		l := Defaultscore()
		b := l.addbuku(buku{Judul: "Buku A", Jumlaheksamplar: 1, JumlahTersedia: 1})

		if err := l.removebuku(b.ID); err != nil {
			t.Fatalf("removebuku error: %v", err)
		}
		if len(l.Buku) != 0 {
			t.Fatalf("jumlah buku = %d, ingin 0", len(l.Buku))
		}
		if err := l.removebuku(99); err == nil {
			t.Fatal("removebuku seharusnya mengembalikan error untuk ID yang tidak ada")
		}
	}

	func TestSearchBukuTidakCaseSensitive(t *testing.T) {
		l := Defaultscore()
		l.addbuku(buku{Judul: "Laut Bercerita"})

		b, err := l.searchbuku("laut")
		if err != nil {
			t.Fatalf("searchbuku error: %v", err)
		}
		if b.Judul != "Laut Bercerita" {
			t.Fatalf("judul = %q", b.Judul)
		}
	}

	func TestTambahJumlahBuku(t *testing.T) {
		l := Defaultscore()
		b := l.addbuku(buku{Judul: "Buku A", Jumlaheksamplar: 1, JumlahTersedia: 1})

		if err := l.tambahjumlahbuku(b.ID, 2); err != nil {
			t.Fatalf("tambahjumlahbuku error: %v", err)
		}
		if b.Jumlaheksamplar != 3 || b.JumlahTersedia != 3 {
			t.Fatalf("stok tidak bertambah dengan benar: %+v", b)
		}
		if err := l.tambahjumlahbuku(b.ID, 0); err == nil {
			t.Fatal("tambahjumlahbuku seharusnya menolak tambahan <= 0")
		}
		if err := l.tambahjumlahbuku(99, 1); err == nil {
			t.Fatal("tambahjumlahbuku seharusnya error untuk ID yang tidak ada")
		}
	}

	func TestPeminjamanDanPengembalian(t *testing.T) {
		l := Defaultscore()
		b := l.addbuku(buku{Judul: "Buku A", Jumlaheksamplar: 1, JumlahTersedia: 1})
		a := l.tambahanggota(anggota{Nama: "Anggota A"})
		now := time.Date(2026, 11, 8, 10, 0, 0, 0, time.Local)

		p, err := l.tambahpeminjaman(b.ID, a.IDAnggota, now)
		if err != nil {
			t.Fatalf("tambahpeminjaman error: %v", err)
		}
		if b.JumlahTersedia != 0 {
			t.Fatalf("stok tersedia = %d, ingin 0", b.JumlahTersedia)
		}
		if !p.Tenggat.Equal(now.AddDate(0, 0, 3)) {
			t.Fatalf("tenggat = %v", p.Tenggat)
		}

		dikembalikan, err := l.kembalikanbuku(p.IDpinjam, now.AddDate(0, 0, 5))
		if err != nil {
			t.Fatalf("kembalikanbuku error: %v", err)
		}
		if dikembalikan.Denda != 2000 {
			t.Fatalf("denda = %d, ingin 2000", dikembalikan.Denda)
		}
		if b.JumlahTersedia != 1 {
			t.Fatalf("stok tersedia = %d, ingin 1", b.JumlahTersedia)
		}
		if _, err := l.kembalikanbuku(p.IDpinjam, now.AddDate(0, 0, 6)); err == nil {
			t.Fatal("buku seharusnya tidak bisa dikembalikan dua kali")
		}
	}

	func TestBatasPeminjaman(t *testing.T) {
		l := Defaultscore()
		a := l.tambahanggota(anggota{Nama: "Anggota A"})
		now := time.Now()

		for i := 0; i < 3; i++ {
			b := l.addbuku(buku{Judul: "Buku", Jumlaheksamplar: 1, JumlahTersedia: 1})
			if _, err := l.tambahpeminjaman(b.ID, a.IDAnggota, now); err != nil {
				t.Fatalf("peminjaman ke-%d error: %v", i+1, err)
			}
		}

		b := l.addbuku(buku{Judul: "Buku keempat", Jumlaheksamplar: 1, JumlahTersedia: 1})
		if _, err := l.tambahpeminjaman(b.ID, a.IDAnggota, now); err == nil {
			t.Fatal("peminjaman keempat seharusnya ditolak")
		}
	}

	func TestPeminjamanDitolakSaatAdaKeterlambatan(t *testing.T) {
		l := Defaultscore()
		a := l.tambahanggota(anggota{Nama: "Anggota A"})
		now := time.Date(2026, 11, 8, 10, 0, 0, 0, time.Local)

		b1 := l.addbuku(buku{Judul: "Buku A", Jumlaheksamplar: 1, JumlahTersedia: 1})
		if _, err := l.tambahpeminjaman(b1.ID, a.IDAnggota, now); err != nil {
			t.Fatalf("peminjaman pertama error: %v", err)
		}

		b2 := l.addbuku(buku{Judul: "Buku B", Jumlaheksamplar: 1, JumlahTersedia: 1})
		if _, err := l.tambahpeminjaman(b2.ID, a.IDAnggota, now.AddDate(0, 0, 4)); err == nil {
			t.Fatal("peminjaman seharusnya ditolak karena ada buku terlambat")
		}
	}

	func TestSaveDataMembuatDirektori(t *testing.T) {
		l := Defaultscore()
		l.addbuku(buku{Judul: "Buku A", Jumlaheksamplar: 1, JumlahTersedia: 1})

		file := filepath.Join(t.TempDir(), "nested", "data.json")
		if err := savedata(file, l); err != nil {
			t.Fatalf("savedata error: %v", err)
		}
		if _, err := os.Stat(file); err != nil {
			t.Fatalf("file tidak dibuat: %v", err)
		}

		hasil, err := loaddata(file)
		if err != nil {
			t.Fatalf("loaddata error: %v", err)
		}
		if len(hasil.Buku) != 1 || hasil.Buku[0].Judul != "Buku A" {
			t.Fatalf("data hasil load tidak sesuai: %+v", hasil.Buku)
		}
	}

	func TestLoadDataMemperbaikiCounterID(t *testing.T) {
		data := &library{
			Buku:            []buku{{ID: 5, Judul: "Buku Lama"}},
			Anggota:         []anggota{{IDAnggota: 7, Nama: "Anggota Lama"}},
			Peminjaman:      []peminjaman{{IDpinjam: 9}},
			NewId:           1,
			NewIDanggota:    1,
			NewIDpeminjaman: 1,
		}

		file := filepath.Join(t.TempDir(), "data.json")
		if err := savedata(file, data); err != nil {
			t.Fatalf("savedata error: %v", err)
		}

		hasil, err := loaddata(file)
		if err != nil {
			t.Fatalf("loaddata error: %v", err)
		}
		if hasil.NewId != 6 {
			t.Fatalf("NewId = %d, ingin 6", hasil.NewId)
		}
		if hasil.NewIDanggota != 8 {
			t.Fatalf("NewIDanggota = %d, ingin 8", hasil.NewIDanggota)
		}
		if hasil.NewIDpeminjaman != 10 {
			t.Fatalf("NewIDpeminjaman = %d, ingin 10", hasil.NewIDpeminjaman)
		}
	}
