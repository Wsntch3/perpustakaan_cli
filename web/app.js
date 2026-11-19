const state = {
  buku: [],
  anggota: [],
  peminjaman: [],
};

const formatRupiah = new Intl.NumberFormat("id-ID", {
  style: "currency",
  currency: "IDR",
  maximumFractionDigits: 0,
});

let timerPesan;

function tampilkanPesan(teks, gagal = false) {
  const el = document.getElementById("pesan");
  el.textContent = teks;
  el.classList.toggle("gagal", gagal);
  el.hidden = false;

  clearTimeout(timerPesan);
  timerPesan = setTimeout(() => {
    el.hidden = true;
  }, 3500);
}

async function panggilApi(url, opsi = {}) {
  const res = await fetch(url, {
    ...opsi,
    headers: opsi.body ? { "Content-Type": "application/json" } : {},
  });

  if (!res.ok) {
    let pesan = "Permintaan gagal";
    try {
      const data = await res.json();
      if (data.error) pesan = data.error;
    } catch (_) {
      // biarkan pesan bawaan
    }
    throw new Error(pesan);
  }

  if (res.status === 204) return null;
  return res.json();
}

function aman(teks) {
  return String(teks ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

function formatTanggal(nilai, denganJam = false) {
  if (!nilai) return "-";
  const tanggal = new Date(nilai);
  if (Number.isNaN(tanggal.getTime())) return "-";

  const opsi = {
    day: "2-digit",
    month: "short",
    year: "numeric",
  };
  if (denganJam) {
    opsi.hour = "2-digit";
    opsi.minute = "2-digit";
  }
  return tanggal.toLocaleString("id-ID", opsi);
}

function renderBuku() {
  const cari = document.getElementById("cari-buku").value.toLowerCase().trim();
  const data = state.buku.filter((b) => b.judul.toLowerCase().includes(cari));
  const tbody = document.getElementById("tabel-buku");

  if (data.length === 0) {
    tbody.innerHTML = '<tr><td colspan="5" class="kosong">Belum ada buku yang cocok.</td></tr>';
    return;
  }

  tbody.innerHTML = data
    .map(
      (b) => `
      <tr>
        <td>${b.id}</td>
        <td>${aman(b.judul)}</td>
        <td>${b.tahun_terbit}</td>
        <td>${b.jumlah_tersedia} dari ${b.jumlah_eksemplar}</td>
        <td class="kolom-aksi">
          <button class="kecil" data-stok="${b.id}" type="button">+ Stok</button>
          <button class="kecil bahaya" data-hapus-buku="${b.id}" type="button">Hapus</button>
        </td>
      </tr>`
    )
    .join("");
}

function renderAnggota() {
  const tbody = document.getElementById("tabel-anggota");

  if (state.anggota.length === 0) {
    tbody.innerHTML = '<tr><td colspan="4" class="kosong">Belum ada anggota.</td></tr>';
    return;
  }

  tbody.innerHTML = state.anggota
    .map(
      (a) => `
      <tr>
        <td>${a.id}</td>
        <td>${aman(a.nama)}</td>
        <td>${formatTanggal(a.tanggal_bergabung)}</td>
        <td class="kolom-aksi">
          <button class="kecil bahaya" data-hapus-anggota="${a.id}" type="button">Hapus</button>
        </td>
      </tr>`
    )
    .join("");
}

function renderPeminjaman() {
  const hanyaAktif = document.getElementById("hanya-aktif").checked;
  const data = state.peminjaman.filter((p) => !hanyaAktif || p.aktif);
  const tbody = document.getElementById("tabel-peminjaman");

  if (data.length === 0) {
    tbody.innerHTML = '<tr><td colspan="8" class="kosong">Belum ada peminjaman.</td></tr>';
    return;
  }

  tbody.innerHTML = data
    .map(
      (p) => `
      <tr>
        <td>${p.id}</td>
        <td>${aman(p.judul_buku || "-")}</td>
        <td>${aman(p.nama_anggota || "-")}</td>
        <td>${formatTanggal(p.tanggal_pinjam, true)}</td>
        <td>${formatTanggal(p.tenggat, true)}</td>
        <td>
          <span class="status ${p.aktif ? "aktif" : "selesai"}">
            ${p.aktif ? "Dipinjam" : "Kembali"}
          </span>
        </td>
        <td>${formatRupiah.format(p.denda)}</td>
        <td class="kolom-aksi">
          ${
            p.aktif
              ? `<button class="kecil" data-kembali="${p.id}" type="button">Kembalikan</button>`
              : ""
          }
        </td>
      </tr>`
    )
    .join("");
}

function isiPilihan() {
  const pilihBuku = document.getElementById("pilih-buku");
  const pilihAnggota = document.getElementById("pilih-anggota");

  pilihBuku.innerHTML =
    '<option value="">Pilih buku</option>' +
    state.buku
      .map(
        (b) =>
          `<option value="${b.id}" ${b.jumlah_tersedia === 0 ? "disabled" : ""}>` +
          `${aman(b.judul)} (${b.jumlah_tersedia} tersedia)</option>`
      )
      .join("");

  pilihAnggota.innerHTML =
    '<option value="">Pilih anggota</option>' +
    state.anggota.map((a) => `<option value="${a.id}">${aman(a.nama)}</option>`).join("");
}

async function muatSemua() {
  try {
    const [buku, anggota, peminjaman] = await Promise.all([
      panggilApi("/api/books"),
      panggilApi("/api/members"),
      panggilApi("/api/loans"),
    ]);

    state.buku = buku;
    state.anggota = anggota;
    state.peminjaman = peminjaman;

    renderBuku();
    renderAnggota();
    renderPeminjaman();
    isiPilihan();
  } catch (err) {
    tampilkanPesan(err.message, true);
  }
}

document.querySelectorAll(".tab").forEach((tombol) => {
  tombol.addEventListener("click", () => {
    document.querySelectorAll(".tab").forEach((t) => {
      t.classList.toggle("active", t === tombol);
    });
    document.querySelectorAll(".panel").forEach((panel) => {
      panel.classList.toggle("active", panel.id === `tab-${tombol.dataset.tab}`);
    });
  });
});

document.getElementById("cari-buku").addEventListener("input", renderBuku);
document.getElementById("hanya-aktif").addEventListener("change", renderPeminjaman);

document.getElementById("form-buku").addEventListener("submit", async (e) => {
  e.preventDefault();
  const form = e.target;

  try {
    await panggilApi("/api/books", {
      method: "POST",
      body: JSON.stringify({
        judul: form.judul.value,
        tahun_terbit: Number(form.tahun_terbit.value),
        jumlah_eksemplar: Number(form.jumlah_eksemplar.value),
      }),
    });
    form.reset();
    await muatSemua();
    tampilkanPesan("Buku berhasil ditambahkan.");
  } catch (err) {
    tampilkanPesan(err.message, true);
  }
});

document.getElementById("tabel-buku").addEventListener("click", async (e) => {
  const tombolStok = e.target.closest("[data-stok]");
  const tombolHapus = e.target.closest("[data-hapus-buku]");

  if (tombolStok) {
    const id = Number(tombolStok.dataset.stok);
    const jawab = prompt("Tambah berapa eksemplar?", "1");
    if (jawab === null) return;

    const tambahan = Number(jawab);
    if (!Number.isInteger(tambahan) || tambahan <= 0) {
      tampilkanPesan("Jumlah stok harus angka lebih dari 0.", true);
      return;
    }

    try {
      await panggilApi(`/api/books/${id}/stock`, {
        method: "POST",
        body: JSON.stringify({ tambahan }),
      });
      await muatSemua();
      tampilkanPesan("Stok buku berhasil ditambahkan.");
    } catch (err) {
      tampilkanPesan(err.message, true);
    }
  }

  if (tombolHapus) {
    const id = Number(tombolHapus.dataset.hapusBuku);
    if (!confirm("Hapus buku ini?")) return;

    try {
      await panggilApi(`/api/books/${id}`, { method: "DELETE" });
      await muatSemua();
      tampilkanPesan("Buku berhasil dihapus.");
    } catch (err) {
      tampilkanPesan(err.message, true);
    }
  }
});

document.getElementById("form-anggota").addEventListener("submit", async (e) => {
  e.preventDefault();
  const form = e.target;

  try {
    await panggilApi("/api/members", {
      method: "POST",
      body: JSON.stringify({ nama: form.nama.value }),
    });
    form.reset();
    await muatSemua();
    tampilkanPesan("Anggota berhasil ditambahkan.");
  } catch (err) {
    tampilkanPesan(err.message, true);
  }
});

document.getElementById("tabel-anggota").addEventListener("click", async (e) => {
  const tombolHapus = e.target.closest("[data-hapus-anggota]");
  if (!tombolHapus) return;

  const id = Number(tombolHapus.dataset.hapusAnggota);
  if (!confirm("Hapus anggota ini?")) return;

  try {
    await panggilApi(`/api/members/${id}`, { method: "DELETE" });
    await muatSemua();
    tampilkanPesan("Anggota berhasil dihapus.");
  } catch (err) {
    tampilkanPesan(err.message, true);
  }
});

document.getElementById("form-peminjaman").addEventListener("submit", async (e) => {
  e.preventDefault();
  const form = e.target;

  try {
    await panggilApi("/api/loans", {
      method: "POST",
      body: JSON.stringify({
        buku_id: Number(form.buku_id.value),
        anggota_id: Number(form.anggota_id.value),
      }),
    });
    form.reset();
    await muatSemua();
    tampilkanPesan("Buku berhasil dipinjam.");
  } catch (err) {
    tampilkanPesan(err.message, true);
  }
});

document.getElementById("tabel-peminjaman").addEventListener("click", async (e) => {
  const tombolKembali = e.target.closest("[data-kembali]");
  if (!tombolKembali) return;

  const id = Number(tombolKembali.dataset.kembali);
  if (!confirm("Kembalikan buku ini?")) return;

  try {
    const hasil = await panggilApi(`/api/loans/${id}/return`, { method: "POST" });
    await muatSemua();

    if (hasil.denda > 0) {
      tampilkanPesan(`Buku dikembalikan. Denda ${formatRupiah.format(hasil.denda)}.`);
    } else {
      tampilkanPesan("Buku dikembalikan tanpa denda.");
    }
  } catch (err) {
    tampilkanPesan(err.message, true);
  }
});

muatSemua();
