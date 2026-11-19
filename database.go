package main

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"

	_ "github.com/lib/pq"
)

var db *sql.DB

const skemaDatabase = `
CREATE TABLE IF NOT EXISTS buku (
	id INTEGER PRIMARY KEY,
	judul TEXT NOT NULL,
	tahun_terbit INTEGER NOT NULL,
	jumlah_eksemplar INTEGER NOT NULL,
	jumlah_tersedia INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS anggota (
	id INTEGER PRIMARY KEY,
	nama TEXT NOT NULL,
	tanggal_bergabung TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS counter (
	nama TEXT PRIMARY KEY,
	nilai INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS peminjaman (
	id INTEGER PRIMARY KEY,
	id_anggota INTEGER NOT NULL REFERENCES anggota(id) ON DELETE CASCADE,
	id_buku INTEGER NOT NULL REFERENCES buku(id) ON DELETE CASCADE,
	tanggal_pinjam TIMESTAMPTZ NOT NULL,
	tenggat TIMESTAMPTZ NOT NULL,
	tanggal_kembali TIMESTAMPTZ,
	denda INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_peminjaman_anggota
	ON peminjaman (id_anggota) WHERE tanggal_kembali IS NULL;
`

func alamatDatabase() string {
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return dsn
	}
	return "postgres://library:library36@localhost:5432/db_library?sslmode=disable"
}

// Tampilkan tujuan koneksi tanpa membocorkan password.
func infoDatabase(dsn string) string {
	u, err := url.Parse(dsn)
	if err != nil || u.Host == "" {
		return "database"
	}
	if u.User != nil && u.User.Username() != "" {
		return fmt.Sprintf("%s@%s%s", u.User.Username(), u.Host, u.Path)
	}
	return u.Host + u.Path
}

func bukaDatabase() error {
	conn, err := sql.Open("postgres", alamatDatabase())
	if err != nil {
		return err
	}

	conn.SetMaxOpenConns(5)
	conn.SetMaxIdleConns(2)

	if err := conn.Ping(); err != nil {
		conn.Close()
		return fmt.Errorf("gagal terhubung ke PostgreSQL: %w", err)
	}

	db = conn
	fmt.Printf("TERHUBUNG KE DATABASE: %s\n", infoDatabase(alamatDatabase()))

	if err := buatTabel(); err != nil {
		conn.Close()
		db = nil
		return fmt.Errorf("gagal menyiapkan tabel: %w", err)
	}

	return nil
}

func tutupDatabase() {
	if db != nil {
		db.Close()
	}
}

func buatTabel() error {
	_, err := db.Exec(skemaDatabase)
	return err
}

func muatData() (*library, error) {
	l := Defaultscore()

	bukuRows, err := db.Query(`
		SELECT id, judul, tahun_terbit, jumlah_eksemplar, jumlah_tersedia
		FROM buku ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer bukuRows.Close()

	for bukuRows.Next() {
		var b buku
		if err := bukuRows.Scan(&b.ID, &b.Judul, &b.TahunTerbit, &b.Jumlaheksamplar, &b.JumlahTersedia); err != nil {
			return nil, err
		}
		l.Buku = append(l.Buku, b)
	}
	if err := bukuRows.Err(); err != nil {
		return nil, err
	}

	anggotaRows, err := db.Query(`
		SELECT id, nama, tanggal_bergabung
		FROM anggota ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer anggotaRows.Close()

	for anggotaRows.Next() {
		var a anggota
		if err := anggotaRows.Scan(&a.IDAnggota, &a.Nama, &a.TanggalBergabung); err != nil {
			return nil, err
		}
		l.Anggota = append(l.Anggota, a)
	}
	if err := anggotaRows.Err(); err != nil {
		return nil, err
	}

	pinjamanRows, err := db.Query(`
		SELECT id, id_anggota, id_buku, tanggal_pinjam, tenggat, tanggal_kembali, denda
		FROM peminjaman ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer pinjamanRows.Close()

	for pinjamanRows.Next() {
		var p peminjaman
		var tanggalKembali sql.NullTime
		if err := pinjamanRows.Scan(
			&p.IDpinjam,
			&p.IDPeminjam,
			&p.IDBuku,
			&p.Tanggalpinjam,
			&p.Tenggat,
			&tanggalKembali,
			&p.Denda,
		); err != nil {
			return nil, err
		}
		if tanggalKembali.Valid {
			p.TanggalKembali = tanggalKembali.Time
		}
		l.Peminjaman = append(l.Peminjaman, p)
	}
	if err := pinjamanRows.Err(); err != nil {
		return nil, err
	}

	l.NewId = bacaCounter("buku", l.NewId)
	l.NewIDanggota = bacaCounter("anggota", l.NewIDanggota)
	l.NewIDpeminjaman = bacaCounter("peminjaman", l.NewIDpeminjaman)
	perbaikiData(l)
	return l, nil
}

func bacaCounter(nama string, cadangan int) int {
	var nilai int
	if err := db.QueryRow(`SELECT nilai FROM counter WHERE nama = $1`, nama).Scan(&nilai); err != nil {
		return cadangan
	}
	return nilai
}

func simpanData(l *library) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`DELETE FROM peminjaman`); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM anggota`); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM buku`); err != nil {
		return err
	}

	for _, b := range l.Buku {
		_, err := tx.Exec(`
			INSERT INTO buku (id, judul, tahun_terbit, jumlah_eksemplar, jumlah_tersedia)
			VALUES ($1, $2, $3, $4, $5)`,
			b.ID, b.Judul, b.TahunTerbit, b.Jumlaheksamplar, b.JumlahTersedia)
		if err != nil {
			return err
		}
	}

	for _, a := range l.Anggota {
		_, err := tx.Exec(`
			INSERT INTO anggota (id, nama, tanggal_bergabung)
			VALUES ($1, $2, $3)`,
			a.IDAnggota, a.Nama, a.TanggalBergabung)
		if err != nil {
			return err
		}
	}

	for _, p := range l.Peminjaman {
		var tanggalKembali interface{}
		if !p.TanggalKembali.IsZero() {
			tanggalKembali = p.TanggalKembali
		}

		_, err := tx.Exec(`
			INSERT INTO peminjaman
				(id, id_anggota, id_buku, tanggal_pinjam, tenggat, tanggal_kembali, denda)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			p.IDpinjam, p.IDPeminjam, p.IDBuku, p.Tanggalpinjam, p.Tenggat, tanggalKembali, p.Denda)
		if err != nil {
			return err
		}
	}

	counters := map[string]int{
		"buku":       l.NewId,
		"anggota":    l.NewIDanggota,
		"peminjaman": l.NewIDpeminjaman,
	}
	for nama, nilai := range counters {
		_, err := tx.Exec(`
			INSERT INTO counter (nama, nilai) VALUES ($1, $2)
			ON CONFLICT (nama) DO UPDATE SET nilai = EXCLUDED.nilai`,
			nama, nilai)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func databaseKosong() (bool, error) {
	var total int
	err := db.QueryRow(`
		SELECT (SELECT count(*) FROM buku)
		     + (SELECT count(*) FROM anggota)
		     + (SELECT count(*) FROM peminjaman)`).Scan(&total)
	if err != nil {
		return false, err
	}
	return total == 0, nil
}
