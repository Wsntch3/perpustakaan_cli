package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

var reader = bufio.NewReader(os.Stdin)


func bacaString(prompt string) string {
	fmt.Print(prompt)
	teks, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF && teks == "" {
			fmt.Println("\nINPUT SELESAI")
			os.Exit(0)
		}
		if err != io.EOF {
			fmt.Println("GAGAL MEMBACA INPUT")
			return ""
		}
	}
	return strings.TrimSpace(teks)
}


func bacaInt(prompt string) int {
	for {
		teks := bacaString(prompt)
		nilai, err := strconv.Atoi(teks)
		if err == nil {
			return nilai
		}
		fmt.Println("INPUT HARUS BERUPA ANGKA")
	}
}

func inoutbuku() buku {
	var b buku

	b.Judul = bacaString("MASUKKAN JUDUL BUKU: ")
	for b.Judul == "" {
		fmt.Println("JUDUL TIDAK BOLEH KOSONG")
		b.Judul = bacaString("MASUKKAN JUDUL BUKU: ")
	}

	b.TahunTerbit = bacaInt("MASUKKAN TAHUN TERBIT BUKU: ")
	for b.TahunTerbit <= 0 {
		fmt.Println("TAHUN TERBIT HARUS LEBIH DARI 0")
		b.TahunTerbit = bacaInt("MASUKKAN TAHUN TERBIT BUKU: ")
	}

	b.Jumlaheksamplar = bacaInt("MASUKKAN BERAPA EKSEMPLAR BUKU: ")
	for b.Jumlaheksamplar <= 0 {
		fmt.Println("JUMLAH EKSEMPLAR HARUS LEBIH DARI 0")
		b.Jumlaheksamplar = bacaInt("MASUKKAN BERAPA EKSEMPLAR BUKU: ")
	}

	b.JumlahTersedia = bacaInt("MASUKKAN JUMLAH BUKU YANG TERSEDIA: ")
	for b.JumlahTersedia < 0 || b.JumlahTersedia > b.Jumlaheksamplar {
		fmt.Println("JUMLAH TERSEDIA HARUS ANTARA 0 DAN JUMLAH EKSEMPLAR")
		b.JumlahTersedia = bacaInt("MASUKKAN JUMLAH BUKU YANG TERSEDIA: ")
	}

	return b
}

func inputanggota() anggota {
	var a anggota

	a.Nama = bacaString("MASUKKAN NAMA ANGGOTA: ")
	for a.Nama == "" {
		fmt.Println("NAMA TIDAK BOLEH KOSONG")
		a.Nama = bacaString("MASUKKAN NAMA ANGGOTA: ")
	}

	return a
}

func menubuku() {
	dataperpus, err := loaddata("data/data.json")
	if err != nil {
		fmt.Println("FILE TIDAK DAPAT DIBUKA:", err)
		return
	}

	for {
		fmt.Println("PILIH LAYANAN PERPUSTAKAAN")
		fmt.Println("1. Tambah Buku")
		fmt.Println("2. Hapus Buku")
		fmt.Println("3. Cari Buku")
		fmt.Println("4. Lihat Semua Buku")
		fmt.Println("5. Tambah Stok Buku")
		fmt.Println("0. Kembali Ke Menu Utama")

		pilih := bacaInt("MASUKKAN PILIHAN JENIS LAYANAN: ")
		switch pilih {
		case 0:
			return
		case 1:
			b := inoutbuku()
			baru := dataperpus.addbuku(b)
			if err := savedata("data/data.json", dataperpus); err != nil {
				fmt.Println("GAGAL MELAKUKAN OPERASI:", err)
				continue
			}
			fmt.Printf("BUKU BERHASIL DITAMBAHKAN DENGAN ID %d\n", baru.ID)
		case 2:
			id := bacaInt("MASUKKAN ID BUKU: ")
			if err := dataperpus.removebuku(id); err != nil {
				fmt.Println("GAGAL MENGHAPUS BUKU:", err)
				continue
			}
			if err := savedata("data/data.json", dataperpus); err != nil {
				fmt.Println("GAGAL MENYIMPAN DATA:", err)
				continue
			}
			fmt.Println("BUKU BERHASIL DIHAPUS")
		case 3:
			judul := bacaString("Masukkan JUDUL BUKU: ")
			if judul == "" {
				fmt.Println("JUDUL TIDAK BOLEH KOSONG")
				continue
			}

			b, err := dataperpus.searchbuku(judul)
			if err != nil {
				fmt.Println("GAGAL MENEMUKAN BUKU:", err)
				continue
			}
			fmt.Printf(
				"BARANG DITEMUKAN:\n"+
					"JUDUL: %s\n"+
					"TAHUN TERBIT: %d\n"+
					"JUMLAH EKSEMPLAR: %d\n"+
					"JUMLAH TERSEDIA: %d\n",
				b.Judul,
				b.TahunTerbit,
				b.Jumlaheksamplar,
				b.JumlahTersedia,
			)
		case 4:
			fmt.Println("DAFTAR BUKU PERPUSTAKAAN")
			b := dataperpus.lihatsemuabuku()
			if len(b) == 0 {
				fmt.Println("BELUM ADA BUKU YANG TERDAFTAR")
				continue
			}
			for _, c := range b {
				fmt.Printf("ID: %d\nJudul: %s\nTahun Terbit: %d\nJumlah Eksamplar: %d\nJumlah Tersedia: %d\n", c.ID, c.Judul, c.TahunTerbit, c.Jumlaheksamplar, c.JumlahTersedia)
			}
		case 5:
			id := bacaInt("MASUKKAN ID: ")
			tambahan := bacaInt("MASUKKAN Tambahan Stok: ")
			if tambahan <= 0 {
				fmt.Println("TAMBAHAN STOK HARUS LEBIH DARI 0")
				continue
			}

			if err := dataperpus.tambahjumlahbuku(id, tambahan); err != nil {
				fmt.Println("GAGAL MENAMBAHKAN TOTAL:", err)
				continue
			}
			if err := savedata("data/data.json", dataperpus); err != nil {
				fmt.Println("GAGAL MENYIMPAN DATA:", err)
				continue
			}
			fmt.Println("STOK BUKU BERHASIL DITAMBAHKAN")
		default:
			fmt.Println("PILIHAN TIDAK SESUAI")
		}
	}
}

func pinjaman() {
	dataperpus, err := loaddata("data/data.json")
	if err != nil {
		fmt.Println("FILE TIDAK DAPAT DIBUKA:", err)
		return
	}

	for {
		fmt.Println("PILIH LAYANAN: ")
		fmt.Println("1. Pinjam Buku")
		fmt.Println("2. Kembalikan Buku")
		fmt.Println("3. Daftar Peminjaman Aktif")
		fmt.Println("0. Exit")

		pilih := bacaInt("MASUKKAN JENIS LAYANAN: ")
		switch pilih {
		case 0:
			return
		case 1:
			idBuku := bacaInt("MASUKKAN ID BUKU: ")
			idAnggota := bacaInt("MASUKKAN ID ANGGOTA: ")

			pinjam, err := dataperpus.tambahpeminjaman(idBuku, idAnggota, time.Now())
			if err != nil {
				fmt.Println("GAGAL MELAKUKAN OPERASI:", err)
				continue
			}
			if err := savedata("data/data.json", dataperpus); err != nil {
				fmt.Println("GAGAL MENYIMPAN DATA:", err)
				continue
			}
			fmt.Printf("BUKU DENGAN ID %d TELAH DIPINJAMKAN\n", pinjam.IDBuku)
			fmt.Printf("ID PEMINJAMAN: %d\n", pinjam.IDpinjam)
			fmt.Printf("TENGGAT PENGEMBALIAN: %s\n", pinjam.Tenggat.Format("2006-01-02 15:04:05"))
		case 2:
			idPinjam := bacaInt("MASUKKAN ID PEMINJAMAN: ")

			pinjam, err := dataperpus.kembalikanbuku(idPinjam, time.Now())
			if err != nil {
				fmt.Println("GAGAL:", err)
				continue
			}
			if err := savedata("data/data.json", dataperpus); err != nil {
				fmt.Println("GAGAL MENYIMPAN DATA:", err)
				continue
			}
			fmt.Printf("BUKU DENGAN ID %d TELAH DIKEMBALIKAN\n", pinjam.IDBuku)
			fmt.Printf("DENDA: Rp%d\n", pinjam.Denda)
		case 3:
			b := dataperpus.daftarpeminjamanaktif()
			if len(b) == 0 {
				fmt.Println("TIDAK ADA PEMINJAMAN AKTIF")
				continue
			}
			for _, p := range b {
				tanggalkembali := "BELUM DIKEMBALIKAN"
				if !p.TanggalKembali.IsZero() {
					tanggalkembali = p.TanggalKembali.Format("2006-01-02 15:04:05")
				}
				fmt.Printf(
					"ID PINJAM: %d\n"+
						"ID ANGGOTA: %d\n"+
						"ID BUKU: %d\n"+
						"TANGGAL PINJAM: %s\n"+
						"TENGGAT: %s\n"+
						"TANGGAL KEMBALI: %s\n"+
						"DENDA: %d\n",
					p.IDpinjam,
					p.IDPeminjam,
					p.IDBuku,
					p.Tanggalpinjam.Format("2006-01-02 15:04:05"),
					p.Tenggat.Format("2006-01-02 15:04:05"),
					tanggalkembali,
					p.Denda,
				)
			}
		default:
			fmt.Println("PILIHAN TIDAK TERSEDIA")
		}
	}
}

func menuanggota() {
	dataperpus, err := loaddata("data/data.json")
	if err != nil {
		fmt.Println("FILE TIDAK DAPAT DIBUKA:", err)
		return
	}

	for {
		fmt.Println("1. Tambah Anggota Perpustakaan")
		fmt.Println("2. Cari Anggota")
		fmt.Println("3. Tampilkan seluruh Anggota")
		fmt.Println("4. Hapus Anggota")
		fmt.Println("0. Kembali Ke Menu Utama")

		pilih := bacaInt("MASUKKAN PILIHAN JENIS LAYANAN: ")
		switch pilih {
		case 0:
			return
		case 1:
			a := inputanggota()
			baru := dataperpus.tambahanggota(a)
			if err := savedata("data/data.json", dataperpus); err != nil {
				fmt.Println("GAGAL MELAKUKAN OPERASI:", err)
				continue
			}
			fmt.Println("DATA BERHASIL TERSIMPAN")
			fmt.Println("Nama:", baru.Nama, "Tanggal Bergabung:", baru.TanggalBergabung.Format("2006-01-02 15:04:05"))
		case 2:
			id := bacaInt("MASUKKAN ID ANGGOTA: ")

			d, err := dataperpus.carianggota(id)
			if err != nil {
				fmt.Println("GAGAL MENEMUKAN ANGGOTA:", err)
				continue
			}
			fmt.Printf(
				"ANGGOTA DITEMUKAN:\n"+
					"ID: %d\n"+
					"NAMA: %s\n"+
					"TANGGAL TERDAFTAR: %s\n",
				d.IDAnggota,
				d.Nama,
				d.TanggalBergabung.Format("2006-01-02 15:04:05"),
			)
		case 3:
			fmt.Println("ANGGOTA PERPUSTAKAAN: ")
			b := dataperpus.tampilkananggota()
			if len(b) == 0 {
				fmt.Println("BELUM ADA ANGGOTA YANG TERDAFTAR")
				continue
			}
			for _, c := range b {
				fmt.Printf("ID: %d\nNAMA: %s\nTANGGAL TERDAFTAR: %s\n", c.IDAnggota, c.Nama, c.TanggalBergabung.Format("2006-01-02 15:04:05"))
			}
		case 4:
			id := bacaInt("MASUKKAN ID: ")

			b, err := dataperpus.hapusanggota(id)
			if err != nil {
				fmt.Println("GAGAL MENGHAPUS ANGGOTA:", err)
				continue
			}
			if err := savedata("data/data.json", dataperpus); err != nil {
				fmt.Println("GAGAL MENYIMPAN DATA:", err)
				continue
			}
			fmt.Println(b.Nama, "TELAH BERHASIL DIHAPUS")
		default:
			fmt.Println("PILIHAN TIDAK SESUAI")
		}
	}
}

func main() {
	for {
		fmt.Println("SILAHKAN PILIH MENU: ")
		fmt.Println("1. MENU BUKU")
		fmt.Println("2. MENU ANGGOTA")
		fmt.Println("3. MENU PINJAM BUKU")
		fmt.Println("0. EXIT")

		pilih := bacaInt("PILIHAN: ")
		switch pilih {
		case 0:
			fmt.Println("PROGRAM SELESAI")
			return
		case 1:
			menubuku()
		case 2:
			menuanggota()
		case 3:
			pinjaman()
		default:
			fmt.Println("PILIHAN TIDAK TERSEDIA")
		}
	}
}
