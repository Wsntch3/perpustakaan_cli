# Perpustakaan CLI

Aplikasi **manajemen perpustakaan berbasis command-line (CLI)** yang dibangun dengan **Go (Golang)**. Project ini mengelola data buku, anggota, serta proses peminjaman dan pengembalian buku secara sederhana, dengan penyimpanan data permanen menggunakan file **JSON**.

Project ini dibuat sebagai implementasi konsep dasar pengembangan perangkat lunak: struktur data, operasi CRUD, validasi aturan bisnis, dan persistensi data — tanpa bergantung pada database eksternal.

---

## Fitur Utama

### Manajemen Buku
- Menambah buku baru dengan judul, tahun terbit, jumlah eksemplar, dan stok tersedia
- Menghapus buku berdasarkan ID
- Mencari buku berdasarkan judul (pencarian sebagian, tidak case-sensitive)
- Menampilkan seluruh daftar buku
- Menambah stok buku yang sudah ada

### Manajemen Anggota
- Mendaftarkan anggota baru (ID dan tanggal bergabung dibuat otomatis)
- Mencari anggota berdasarkan ID
- Menampilkan seluruh anggota
- Menghapus anggota berdasarkan ID

### Peminjaman & Pengembalian
- Meminjam buku dengan validasi:
  - Buku harus tersedia
  - Maksimal **3 buku aktif** per anggota
  - Anggota tidak boleh memiliki pinjaman yang terlambat
- Tenggat peminjaman otomatis **3 hari** setelah tanggal pinjam
- Pengembalian buku dengan perhitungan denda otomatis
- Menampilkan daftar peminjaman yang masih aktif
- Stok buku otomatis berkurang saat dipinjam dan bertambah saat dikembalikan

### Persistensi Data
- Seluruh data disimpan otomatis ke `data/data.json`
- ID buku, anggota, dan peminjaman menggunakan auto-increment
- Counter ID otomatis disinkronkan dengan data yang sudah ada saat file dimuat
- Data tetap tersimpan meskipun program ditutup

### Validasi & Ketahanan
- Input angka divalidasi dan diulang jika pengguna memasukkan bukan angka
- Judul buku dan nama anggota tidak boleh kosong
- Jumlah stok tidak boleh melebihi jumlah eksemplar
- Folder `data/` dibuat otomatis jika belum tersedia
- Operasi yang gagal tidak menghentikan program (tidak ada panic)
- Business logic inti dilindungi unit test

---

## Teknologi yang Digunakan

| Komponen | Keterangan |
| --- | --- |
| Bahasa | Go (Golang) 1.26+ |
| Penyimpanan | File JSON lokal (`data/data.json`) |
| Antarmuka | Terminal / Command Line |
| Library | Hanya standard library (`encoding/json`, `os`, `bufio`, `time`, dll.) |

Project ini sengaja tidak menggunakan framework atau dependency eksternal agar fokus pada logika dan struktur program.

---

## Struktur Project

```
perpustakaan_cli/
├── main.go        # Entry point, menu utama, dan handler input pengguna
├── library.go     # Struct inti & seluruh business logic perpustakaan
├── book.go        # Model/struct buku
├── anggota.go     # Model/struct anggota
├── borrow.go      # Model/struct peminjaman
├── file.go        # Load & save data JSON
├── library_test.go # Unit test untuk business logic inti
├── go.mod         # Metadata module Go
└── data/
    └── data.json  # Penyimpanan data aplikasi
```

### Penjelasan Singkat

- **`main.go`** — Menjalankan program, menampilkan menu, dan menghubungkan input pengguna dengan fungsi pada `library.go`.
- **`library.go`** — Pusat logika aplikasi: CRUD buku & anggota, validasi peminjaman, perhitungan denda, dan pengembalian buku.
- **`file.go`** — Membaca dan menulis data ke JSON. Jika file belum ada, aplikasi otomatis membuat data awal (default).
- **`library_test.go`** — Unit test untuk CRUD, validasi peminjaman, perhitungan denda, dan persistensi data.
- **`book.go`, `anggota.go`, `borrow.go`** — Definisi entitas beserta tag JSON untuk serialisasi.

---

## Model Data

### Buku

| Field | Tipe | Deskripsi |
| --- | --- | --- |
| `ID` | `int` | ID unik buku |
| `Judul` | `string` | Judul buku |
| `TahunTerbit` | `int` | Tahun terbit |
| `Jumlaheksamplar` | `int` | Total eksemplar yang dimiliki |
| `JumlahTersedia` | `int` | Stok yang sedang tersedia |

### Anggota

| Field | Tipe | Deskripsi |
| --- | --- | --- |
| `IDAnggota` | `int` | ID unik anggota |
| `Nama` | `string` | Nama anggota |
| `TanggalBergabung` | `time.Time` | Waktu pendaftaran |

### Peminjaman

| Field | Tipe | Deskripsi |
| --- | --- | --- |
| `IDpinjam` | `int` | ID unik transaksi peminjaman |
| `IDPeminjam` | `int` | ID anggota peminjam |
| `IDBuku` | `int` | ID buku yang dipinjam |
| `Tanggalpinjam` | `time.Time` | Waktu peminjaman |
| `Tenggat` | `time.Time` | Batas waktu pengembalian |
| `TanggalKembali` | `time.Time` | Waktu pengembalian (zero value jika belum) |
| `Denda` | `int` | Denda keterlambatan (Rupiah) |

---

## Aturan Bisnis

1. Setiap anggota maksimal meminjam **3 buku aktif** secara bersamaan.
2. Anggota yang masih memiliki buku terlambat **tidak dapat meminjam** buku baru.
3. Buku hanya dapat dipinjam jika `JumlahTersedia > 0`.
4. Tenggat pengembalian adalah **3 hari** setelah tanggal pinjam.
5. Denda keterlambatan dihitung **Rp1.000 per hari**.
6. Buku yang sudah dikembalikan tidak dapat dikembalikan dua kali.
7. Stok tersedia tidak akan melebihi jumlah eksemplar.

---

## Cara Menjalankan

### Prasyarat

- Go **1.26 atau lebih baru** sudah terpasang.
  Cek dengan:

  ```bash
  go version
  ```

### Menjalankan Langsung

```bash
git clone <URL_REPOSITORY>
cd perpustakaan_cli
go run .
```

### Build Menjadi Binary

```bash
go build -o perpustakaan
./perpustakaan
```

> Folder `data/` dan file `data/data.json` akan dibuat otomatis saat data pertama kali disimpan, sehingga aplikasi tetap bisa dijalankan dari kondisi bersih.

---

## 🧪 Pengujian

Project ini memiliki unit test untuk memverifikasi business logic inti, termasuk regresi bug pada pengelolaan anggota dan persistensi data.

```bash
go test ./... -v
```

Cakupan pengujian meliputi:

- Penambahan, pencarian, dan penghapusan buku
- Penambahan dan penghapusan anggota (termasuk anggota terakhir)
- Validasi stok dan batas peminjaman
- Peminjaman, pengembalian, dan perhitungan denda
- Penyimpanan/muat data JSON serta sinkronisasi counter ID

---

## 🖥️ Contoh Penggunaan

Saat program dijalankan, pengguna akan disajikan menu utama:

```text
SILAHKAN PILIH MENU:
1. MENU BUKU
2. MENU ANGGOTA
3. MENU PINJAM BUKU
0. EXIT
PILIHAN:
```

### Alur Singkat

1. Pilih **Menu Buku** → tambahkan buku beserta stoknya.
2. Pilih **Menu Anggota** → daftarkan anggota baru dan catat ID-nya.
3. Pilih **Menu Pinjam Buku** → masukkan ID buku dan ID anggota.
4. Untuk pengembalian, pilih menu **Kembalikan Buku**, lalu masukkan **ID peminjaman**.
5. Gunakan menu **Daftar Peminjaman Aktif** untuk melihat status pinjaman, tenggat, dan denda.

---

## 🏗️ Alur Arsitektur

```text
┌─────────────────────┐
│      main.go        │
│  Menu & Input CLI   │
└──────────┬──────────┘
           │ memanggil
           ▼
┌─────────────────────┐        ┌──────────────────┐
│     library.go      │◄──────►│  Model (struct)  │
│   Business Logic    │        │ book / anggota / │
│  CRUD & Validasi    │        │    peminjaman    │
└──────────┬──────────┘        └──────────────────┘
           │ load / save
           ▼
┌─────────────────────┐
│      file.go        │
│  JSON Persistence   │
└──────────┬──────────┘
           ▼
     data/data.json
```

Pemisahan ini membuat kode lebih mudah dibaca, diuji, dan dikembangkan — misalnya ketika nanti ingin menambahkan REST API atau antarmuka GUI.

---

## 🎯 Tujuan Project

Project ini dibuat untuk melatih dan menunjukkan kemampuan dalam:

- Menulis kode Go yang terstruktur dan modular
- Menerapkan operasi CRUD dan validasi aturan bisnis
- Mengelola state dan persistensi data tanpa database
- Merancang antarmuka CLI yang mudah digunakan
- Memisahkan model, logika bisnis, dan lapisan penyimpanan




