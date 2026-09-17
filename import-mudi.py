#!/usr/bin/env python3
"""Import CSV Google Form MUDI Semampir → members table.

Usage:
  python3 import-mudi.py data/mudi-semampir.csv > /tmp/import.sql
  docker compose exec -T postgres psql -U pengajian -d pengajian < /tmp/import.sql
"""

import csv
import re
import sys
import uuid

BULAN = {
    # Indonesia
    "januari": 1, "februari": 2, "maret": 3, "april": 4,
    "mei": 5, "juni": 6, "juli": 7, "agustus": 8,
    "september": 9, "oktober": 10, "november": 11, "desember": 12,
    # Singkatan Indonesia
    "jan": 1, "feb": 2, "mar": 3, "apr": 4,
    "jun": 6, "jul": 7, "agu": 8, "ags": 8,
    "sep": 9, "sept": 9, "okt": 10, "nov": 11, "des": 12,
    # English (form pakai Google Form default)
    "january": 1, "february": 2, "march": 3, "april": 4,
    "may": 5, "june": 6, "july": 7, "august": 8,
    "sept": 9, "october": 10, "november": 11, "december": 12,
    "aug": 8, "oct": 10, "dec": 12,
}


def new_id(prefix):
    return prefix + uuid.uuid4().hex[:8].upper()


def parse_tgl(s):
    """Parse tanggal dari berbagai format."""
    s = (s or "").strip()
    if not s:
        return None

    # Fix typo tahun
    s = re.sub(r"\b(19\d{3,}|20\d{3,})\b", lambda m: m.group(1)[:4], s)

    # Normalisasi: replace separator dengan spasi (kecuali digit-to-alpha boundary)
    # "23maret" → "23 maret"
    s_norm = re.sub(r"(\d)([a-zA-Z])", r"\1 \2", s)
    s_norm = re.sub(r"([a-zA-Z])(\d)", r"\1 \2", s_norm)
    # Replace / - . , dengan spasi
    s_norm = re.sub(r"[/,.\-]", " ", s_norm)
    s_norm = re.sub(r"\s+", " ", s_norm).strip()

    parts = s_norm.split()

    # Cari pattern: [day] [month_name] [year]
    for i in range(len(parts) - 2):
        try:
            day = int(parts[i])
        except ValueError:
            continue
        month_str = parts[i + 1].lower()
        if month_str not in BULAN:
            continue
        try:
            year = int(parts[i + 2])
        except (ValueError, IndexError):
            continue
        if year < 100:
            year += 2000
        if 1 <= day <= 31 and 1900 < year < 2100:
            return f"{year:04d}-{BULAN[month_str]:02d}-{day:02d}"

    # Cari pattern: [day] [month_num] [year] (semua angka)
    nums = []
    for p in parts:
        try:
            nums.append(int(p))
        except ValueError:
            continue
    if len(nums) >= 3:
        day, month, year = nums[0], nums[1], nums[2]
        if year < 100:
            year += 2000
        if 1 <= day <= 31 and 1 <= month <= 12 and 1900 < year < 2100:
            return f"{year:04d}-{month:02d}-{day:02d}"

    return None


def parse_tempat_lahir(s):
    s = (s or "").strip()
    if "," in s:
        return s.split(",", 1)[0].strip()
    # Kalau isinya cuma tanggal (tidak ada koma), kembalikan kosong
    return ""


def normalize_gender(s):
    s = (s or "").strip().lower()
    if "perempuan" in s or "wanita" in s or s == "p":
        return "P"
    if "laki" in s or "pria" in s or s == "l":
        return "L"
    return ""


def normalize_wa(s):
    if not s:
        return ""
    # Kalau ada multiple nomor, ambil yang pertama
    s = re.split(r"[/]", s)[0]
    digits = re.sub(r"\D", "", s)
    if not digits:
        return ""
    if digits.startswith("0"):
        digits = "62" + digits[1:]
    elif not digits.startswith("62"):
        digits = "62" + digits
    return digits


def convert_foto(s):
    s = (s or "").strip()
    if not s:
        return ""
    m = re.search(r"[?&]id=([^&]+)", s)
    if m:
        return f"https://drive.google.com/thumbnail?id={m.group(1)}&sz=w1000"
    return s


def map_pekerjaan(s):
    s = (s or "").strip()
    low = s.lower()
    if "mahasiswa" in low:
        return "Mahasiswa"
    if "sekolah" in low or "siswa" in low or "pelajar" in low:
        return "Pelajar"
    if "bekerja" in low:
        return "Bekerja"
    if "usaha" in low:
        return "Wiraswasta"
    if "magang" in low:
        return "Magang"
    if "ngajar" in low or "guru" in low:
        return "Guru"
    if "bantu" in low:
        return "Bantu orang tua"
    return s


def esc(v):
    """Escape string untuk SQL. Empty string → NULL (untuk kolom nullable DATE)."""
    if v is None:
        return "NULL"
    v = str(v).strip()
    if not v:
        return "NULL"
    return "'" + v.replace("'", "''") + "'"


def esc_str(v):
    """Escape string untuk kolom TEXT NOT NULL DEFAULT ''. Empty → ''."""
    if v is None:
        return "''"
    v = str(v).strip()
    return "'" + v.replace("'", "''") + "'"


def get_field(row, name):
    """Fuzzy lookup header dengan strip."""
    for k, v in row.items():
        if k and k.strip().startswith(name):
            return (v or "").strip()
    return ""


def main():
    if len(sys.argv) < 2:
        print("Usage: python3 import-mudi.py <csv-file>", file=sys.stderr)
        sys.exit(1)

    with open(sys.argv[1], encoding="utf-8") as f:
        reader = csv.DictReader(f)
        print("BEGIN;")
        count = 0

        for row in reader:
            nama = get_field(row, "Nama Lengkap")
            if not nama or len(nama) < 2:
                continue

            jk = normalize_gender(get_field(row, "Jenis Kelamin"))
            tempat_lahir_raw = get_field(row, "Tempat Lahir")
            tgl_lahir_raw = get_field(row, "Tanggal Lahir")

            # Gabung dua kolom supaya tanggal yang terpisah tetap kedetect
            combined = ((tempat_lahir_raw or "") + " " + (tgl_lahir_raw or "")).strip()
            tgl = parse_tgl(combined)
            tempat = parse_tempat_lahir(tempat_lahir_raw)

            no_wa = normalize_wa(get_field(row, "Nomor HP / WA"))
            alamat = get_field(row, "Alamat di Surabaya")
            hobi = get_field(row, "Hobi")
            status_raw = get_field(row, "Status")
            is_muballigh = (
                "mubaligh" in status_raw.lower()
                and "non" not in status_raw.lower()
            )
            aktifitas = get_field(row, "Aktifitas / Kegiatan")
            pekerjaan = map_pekerjaan(aktifitas)
            universitas = get_field(row, "Nama Universitas")
            jurusan = get_field(row, "Jurusan")
            tahun = get_field(row, "Tahun Angkatan")
            foto = convert_foto(get_field(row, "Pas Photo"))
            kota_asal = get_field(row, "Kota Asal")
            nama_panggilan = get_field(row, "Nama Panggilan")

            # Skip baris yang jelas dummy (test form)
            if nama.lower() in ("jjj", "hgh") or "test" in nama.lower():
                print(f"-- SKIP dummy: {nama}", file=sys.stderr)
                continue

            member_id = new_id("MBR")
            is_kerja = "true" if pekerjaan == "Bekerja" else "false"
            muballigh_sql = "true" if is_muballigh else "false"

            sql = (
                "INSERT INTO members ("
                "member_id, nama_lengkap, nama_panggilan, jenis_kelamin, "
                "tempat_lahir, tanggal_lahir, foto_url, no_wa, "
                "alamat_rumah, desa, daerah, kelompok, "
                "is_muballigh, is_kerja, is_nikah, hobi, pekerjaan, "
                "status_pembinaan, status_aktif, tanggal_masuk, "
                "jenjang_pendidikan, sekolah, jurusan, "
                "tahun_mulai_pendidikan, tahun_selesai_pendidikan, "
                "created_at, updated_at"
                ") VALUES ("
                f"{esc(member_id)}, {esc_str(nama)}, {esc_str(nama_panggilan)}, {esc_str(jk)}, "
                f"{esc_str(tempat)}, {esc(tgl)}, {esc_str(foto)}, {esc_str(no_wa)}, "
                f"{esc_str(alamat)}, 'Medokan Semampir', {esc_str(kota_asal)}, 'Semampir', "
                f"{muballigh_sql}, {is_kerja}, false, {esc_str(hobi)}, {esc_str(pekerjaan)}, "
                f"'AKTIF', true, CURRENT_DATE, "
                f"'', {esc_str(universitas)}, {esc_str(jurusan)}, "
                f"{esc_str(tahun)}, '', "
                f"NOW(), NOW()"
                ") ON CONFLICT DO NOTHING;"
            )
            print(sql)
            count += 1

        print(f"-- Total: {count} rows")
        print("COMMIT;")


if __name__ == "__main__":
    main()
