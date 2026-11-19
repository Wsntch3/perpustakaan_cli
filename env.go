package main

import (
	"bufio"
	"os"
	"strings"
)

// bacaEnv membaca file .env jika ada. Environment yang sudah di-set
// dari terminal tidak ditimpa.
func bacaEnv(file string) {
	f, err := os.Open(file)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		baris := strings.TrimSpace(scanner.Text())
		if baris == "" || strings.HasPrefix(baris, "#") {
			continue
		}

		kunci, nilai, ok := strings.Cut(baris, "=")
		if !ok {
			continue
		}

		kunci = strings.TrimSpace(kunci)
		nilai = strings.Trim(strings.TrimSpace(nilai), `"'`)

		if _, sudahAda := os.LookupEnv(kunci); !sudahAda {
			os.Setenv(kunci, nilai)
		}
	}
}
