# Datalogger Analysis Application — Customer User Manual
**User Guide for Environmental Data Monitoring, Telemetry Analysis and Reporting**

- **Document ID:** DAA-CUM-001
- **Versi Dokumen:** 1.0
- **Status Dokumen:** Draft for Customer Review
- **Klasifikasi:** Customer Documentation
- **Penyusun (Prepared by):** [Nama Perusahaan Penyedia Sistem]
- **Pelanggan (Customer):** [Nama Perusahaan / Instansi Customer]
- **Lokasi / Stasiun:** [Nama Proyek atau Lokasi Stasiun Monitoring]
- **Format Dokumen Resmi:** Tersedia dalam format Word DOCX siap cetak & edit: `Datalogger_Analysis_Application_Customer_User_Manual.docx`

---

## Lembar Pengesahan Dokumen

| Peran / Kewenangan | Nama Lengkap & Gelar | Tanda Tangan | Tanggal |
| :--- | :--- | :---: | :---: |
| **Disusun Oleh (Prepared by)**<br>Project Engineer / Tech Writer | [Nama Penyusun] | _________________ | [DD/MM/YYYY] |
| **Ditinjau Oleh (Reviewed by)**<br>Lead Systems Engineer | [Nama Reviewer] | _________________ | [DD/MM/YYYY] |
| **Disetujui Penyedia (Approved by)**<br>Project Director / Manager | [Nama Direktur/PM] | _________________ | [DD/MM/YYYY] |
| **Diterima Customer (Accepted by)**<br>Kepala Seksi / Site Manager | [Nama Pejabat Customer] | _________________ | [DD/MM/YYYY] |

---

## Riwayat Perubahan Dokumen (Revision History)

| Versi | Tanggal | Penyusun / Editor | Deskripsi Perubahan |
| :---: | :---: | :--- | :--- |
| **1.0** | [Tanggal Rilis] | [Nama Perusahaan] | Penerbitan perdana dokumen Customer User Manual serah terima sistem stasiun AQMS. |

---

## Daftar Isi
1. [Bab 1 — Pengenalan Aplikasi](#bab-1--pengenalan-aplikasi)
2. [Bab 2 — Persiapan Menggunakan Aplikasi](#bab-2--persiapan-menggunakan-aplikasi)
3. [Bab 3 — Login dan Logout Akun](#bab-3--login-dan-logout-akun)
4. [Bab 4 — Mengenal Dashboard Operator (SCADA Console)](#bab-4--mengenal-dashboard-operator-scada-console)
5. [Bab 5 — Memantau Data Sensor Lingkungan](#bab-5--memantau-data-sensor-lingkungan)
6. [Bab 6 — Melihat Grafik dan Analisis Trend](#bab-6--melihat-grafik-dan-analisis-trend)
7. [Bab 7 — Pemantauan Kepatuhan Regulasi ISPU (Permen LHK No. 14/2020)](#bab-7--pemantauan-kepatuhan-regulasi-ispu-permen-lhk-no-142020)
8. [Bab 8 — Memahami Struktur Data: Raw, Processed, dan Aggregated](#bab-8--memahami-struktur-data-raw-processed-dan-aggregated)
9. [Bab 9 — Manajemen Alarm dan Notifikasi Insiden](#bab-9--manajemen-alarm-dan-notifikasi-insiden)
10. [Bab 10 — Export Data dan Pengunduhan Laporan](#bab-10--export-data-dan-pengunduhan-laporan)
11. [Bab 11 — Pengaturan Pengguna dan Personalisasi](#bab-11--pengaturan-pengguna-dan-personalisasi)
12. [Bab 12 — Panduan Troubleshooting Praktis Customer](#bab-12--panduan-troubleshooting-praktis-customer)
13. [Bab 13 — Prosedur Standar Operasional Harian (Daily SOP)](#bab-13--prosedur-standar-operasional-harian-daily-sop)
14. [Bab 14 — Layanan Dukungan Teknis dan Eskalasi Vendor](#bab-14--layanan-dukungan-teknis-dan-eskalasi-vendor)
15. [Lampiran-Lampiran Lengkap](#lampiran-lampiran-lengkap)

---

## Bab 1 — Pengenalan Aplikasi

### 1.1 Apa itu Datalogger Analysis Application?
**Datalogger Analysis Application** adalah platform Supervisory Control and Data Acquisition (SCADA) dan telemetri lingkungan industri yang dirancang untuk akuisisi, validasi, penghitungan baku mutu, dan pelaporan data pemantauan kualitas udara (*Air Quality Monitoring System* / AQMS) serta stasiun meteorologi secara kontinu 24 jam non-stop.

Sistem bekerja di tingkat *edge* (*Edge Computing Platform*) yang terhubung langsung dengan sensor industri stasiun melalui protokol Modbus RTU (serial RS-485) dan Modbus TCP (Ethernet). Data mentah diverifikasi kualitasnya, dihitung sesuai regulasi nasional, dan disimpan secara aman pada MariaDB Edge lokal di stasiun.

### 1.2 Tujuan dan Manfaat bagi Customer
1. **Visibilitas Real-Time Penuh:** Memantau seluruh parameter gas pencemar dan cuaca detik-demi-detik melalui antarmuka web responsif.
2. **Kepatuhan Mutlak Regulasi ISPU:** Perhitungan otomatis Indeks Standar Pencemar Udara mengacu pada **Permen LHK No. 14 Tahun 2020**.
3. **Integritas Data Terjamin:** Algoritma otonom mendeteksi sensor macet (*stuck value*), lonjakan palsu (*spike*), nilai di luar rentang (*out-of-range*), dan putus kabel.
4. **Peringatan Dini Cepat:** Notifikasi alarm visual dan suara saat konsentrasi gas melampaui batas ambang bahaya.
5. **Kemudahan Audit & Pelaporan:** Cetak instan PDF, ekspor spreadsheet CSV Excel, dan berkas mandiri HTML.

### 1.3 Jenis Parameter yang Dipantau
- **Gas Pencemar & Partikulat (`AQMS-02`):** PM 2.5, PM 10, SO2, NO2, CO, O3, dan HC dalam satuan $\mu\text{g/m}^3$.
- **Meteorologi & Cuaca (`AQMS-01`):** Suhu Udara (°C), Kelembaban Relatif (%), Tekanan Barometrik (hPa), Kecepatan Angin (m/s), Arah Angin (°), Radiasi Matahari (W/m²), Curah Hujan (mm).

### 1.4 Alur Data Sensor Menuju Dashboard SCADA
```
[Sensor & Analyzer Fisik]
        │ (Modbus RTU / TCP)
        ▼
[Datalogger Edge Engine] ───► [Quality & Scaling Engine]
                                      │ (Validasi GOOD/UNCERTAIN/BAD)
                                      ▼
                             [MariaDB Edge Local]
                                      │
        ┌─────────────────────────────┴─────────────────────────────┐
        ▼                                                           ▼
[Web SCADA Dashboard]                                   [Outbound Forwarder]
 (Operator Console, ISPU, Trends)                        (KLHK / SISPEK / MQTT)
```

---

## Bab 2 — Persiapan Menggunakan Aplikasi

### 2.1 Kebutuhan Workstation Pengguna
- **CPU:** Dual-Core 2.0 GHz (Min), Quad-Core 2.8 GHz (Rekomendasi).
- **RAM:** 4 GB (Min), 8 GB atau lebih (Rekomendasi).
- **Layar:** Resolusi minimum $1366 \times 768$, rekomendasi Full HD $1920 \times 1080$.
- **Jaringan:** LAN Ethernet 100/1000 Mbps atau Wi-Fi industri terproteksi.

### 2.2 Peramban (Browser) yang Didukung
- **Google Chrome** (Versi 110 ke atas) — *Sangat Direkomendasikan*.
- **Microsoft Edge** (Versi 110 ke atas).
- **Mozilla Firefox** (Versi 115 ESR / 120 ke atas).
*(Internet Explorer tidak didukung).*

### 2.3 Alamat URL Akses Aplikasi
Akses web aplikasi melalui browser:
```text
http://[IP_ADDRESS_DATALOGGER]:8080/#/login
Contoh: http://192.168.1.53:8080/#/login
```

---

## Bab 3 — Login dan Logout Akun

### 3.1 Antarmuka Login
Halaman login memuat field **USERNAME**, **PASSWORD**, tombol **Sign In →**, pengalih bahasa **EN | ID**, dan pengalih tema **Dark / Light Mode**.

### 3.2 Struktur Peran dan Hak Akses (RBAC)
- **`OPERATOR`:** Akses SCADA Operator Console, pemantauan sensor, visualisasi ISPU, penerimaan alarm, dan ekspor laporan.
- **`ENGINEER`:** Semua hak Operator ditambah konfigurasi perangkat sensor, formula kalibrasi, pengujian pembacaan register (*Test Read*), dan pemantau frame komunikasi.
- **`ADMIN`:** Semua hak Engineer ditambah pendaftaran dan pengelolaan akun operator lokal customer, backup & restore, dan log sistem.
- **`SUPERADMIN`:** Hak tertinggi teknisi vendor untuk pemeliharaan inti engine dan database.

### 3.3 Penanganan Layar HTTP 403 Forbidden
Jika akun role Operator mencoba mengakses menu khusus Admin (misal `/administration/users`), sistem akan menampilkan layar pengamanan **"Access Restricted (HTTP 403)"**. Klik tombol biru **Operator Console** untuk kembali ke halaman kerja Anda.

---

## Bab 4 — Mengenal Dashboard Operator (SCADA Console)

Halaman utama **Operator Command Center** (`/#/monitoring/operator`) menyajikan 6 zona informasi:
1. **Header Terpadu:** Logo Instansi Customer, Jam Digital Pusat presisi, dan Logo Penyedia.
2. **SCADA Toolbar:** Tombol Auto-Hide Sidebar, Fullscreen toggle, dan opsi filter tampilan grid.
3. **Matriks Kartu Sensor:** 14 kartu parameter real-time dengan nilai numerik besar, satuan, tag kategori (GAS / MET), dan sparkline mini.
4. **Strip Status ISPU Cepat:** Menampilkan nilai ISPU stasiun, kategori (misal: SEDANG), parameter kritis dominan (misal: PM25), dan persentase kelengkapan data regulasi valid.
5. **Grafik Trend Multi-Parameter Bawah:** Kurva interaktif 14 parameter bersamaan dengan pilihan rentang waktu (1m, 5m, 15m, 1h, 6h).
6. **Bilah Aksi Cepat:** Pintasan Export Data, Generate Report, Alarm History, dan Sound Alert toggle.

---

## Bab 5 — Memantau Data Sensor Lingkungan

### 5.1 Menu Devices & Sensors (`/#/monitoring/devices`)
Menampilkan ringkasan status peralatan fisik:
- **Total Registered:** 2 unit (AQMS-01 dan AQMS-02).
- **Status Komunikasi:** ONLINE (Hijau), OFFLINE (Merah), MAINTENANCE (Kuning).
- **Waktu Data Terakhir (Last Seen / Last Data):** Timestamp penerimaan data terkini.

### 5.2 Menu Live Telemetry (`/#/monitoring/telemetry`)
Menampilkan aliran data langsung dengan status mutu:
- **`GOOD` (Hijau):** Data valid, CRC Modbus benar, nilai berada di rentang wajar.
- **`UNCERTAIN` (Kuning):** Nilai mendekati batas ambang atau sensor terindikasi stuck.
- **`BAD` (Merah):** Putus komunikasi, nilai di luar rentang fisik, atau terjadi spike.

---

## Bab 6 — Melihat Grafik dan Analisis Trend

Pada modul **Unified Telemetry Analysis** (`/#/monitoring/analysis`):
1. **Pilihan Mode Data:**
   - *Historical / Processed:* Data kalibrasi seketika.
   - *Aggregation & Rollup:* Rata-rata periodik (1m, 5m, 1h, 1d) untuk evaluasi jangka panjang.
2. **Pilihan Skala Grafik:**
   - *Normalized (0–100%):* Menstandarkan kurva seluruh parameter untuk melihat korelasi naik-turun antar gas dan cuaca.
   - *Native Scale:* Menampilkan nilai fisik asli sesuai satuan masing-masing.
3. **Interaksi Tooltip & Zoom:** Arahkan kursor ke garis grafik untuk membaca timestamp presisi dan nilai seluruh sensor.

---

## Bab 7 — Pemantauan Kepatuhan Regulasi ISPU (Permen LHK No. 14/2020)

### 7.1 Dua Mode Perhitungan ISPU
- **Mode 24-Jam (Regulasi):** Menghitung rata-rata bergerak 24 jam (*24-Hour Rolling Average*) sesuai formula Lampiran II Permen LHK 14/2020. Mode ini merupakan acuan resmi pelaporan.
- **Mode Per Agregasi (Operasional):** Menghitung indeks real-time untuk mendeteksi perubahan polusi mendadak.

### 7.2 Skala & Kategori Mutu Udara ISPU
| Rentang Indeks | Kategori ISPU | Warna Resmi | Dampak Terhadap Kesehatan |
| :---: | :---: | :---: | :--- |
| **0 – 50** | **BAIK** | Hijau | Sangat baik, tidak ada dampak negatif pada manusia atau hewan. |
| **51 – 100** | **SEDANG** | Biru | Mutu udara dapat diterima pada kesehatan manusia secara umum. |
| **101 – 200** | **TIDAK SEHAT** | Kuning | Merugikan manusia, hewan, dan tumbuhan yang sensitif. |
| **201 – 300** | **SANGAT TIDAK SEHAT** | Merah | Meningkatkan risiko kesehatan pada segmen populasi terpapar. |
| **> 300** | **BERBAHAYA** | Hitam / Ungu | Berbahaya serius bagi seluruh populasi; peringatan darurat. |

### 7.3 Syarat Sah Regulasi (Kelengkapan Data $\ge 75\%$)
Sesuai Pasal 7 ayat (3) Permen LHK 14/2020, indeks ISPU 24 jam hanya berstatus **VALID** apabila data stasiun terisi minimal 75% (minimal 18 jam data valid dalam sehari).

---

## Bab 8 — Memahami Struktur Data: Raw, Processed, dan Aggregated

- **Raw Data:** Byte heksadesimal murni dari register Modbus sensor (jejak audit integritas fisik).
- **Processed Data:** Nilai Float32 hasil perkalian faktor skala kalibrasi dan validasi kualitas.
- **Aggregated Data (Rollup):** Rata-rata berkala (1m, 1h, 1d) yang disimpan untuk efisiensi penyimpanan dan tren historis.
- *Catatan:* Perbedaan antara nilai kartu sensor (sesaat/instantaneous) dengan kurva grafik adalah hal wajar karena grafik menampilkan rata-rata representatif dalam interval rollup.

---

## Bab 9 — Manajemen Alarm dan Notifikasi Insiden

### 9.1 Tingkat Keparahan (Severity)
- **`CRITICAL` (Merah):** Melampaui baku mutu kritis, sensor mati total, atau kegagalan transmisi.
- **`WARNING` (Kuning):** Mendekati ambang batas, sensor macet (*stuck value*), atau latensi komunikasi naik.
- **`INFO` (Biru):** Pemberitahuan sistem, pergantian mode, atau selesai backup.

### 9.2 Prosedur Operator Menangani Alarm
1. Buka menu **Active Incidents** (`/#/alarm/active`).
2. Identifikasi instrumen dan parameter penyebab pemicu alarm.
3. Klik tombol **Acknowledge** untuk mengonfirmasi bahwa operator sedang menangani insiden.
4. Catat waktu dan nomor kejadian pada buku log operasional shift.
5. Jika alarm kritis berlangsung > 15 menit, lakukan eskalasi ke tim teknis vendor.

---

## Bab 10 — Export Data dan Pengunduhan Laporan

Pada menu **Reports** (`/#/monitoring/reports`):
1. **Jenis Laporan:** *Comprehensive Operational*, *Telemetry & Quality Focus*, *Alarms & Events Audit*, dan *Hardware Availability*.
2. **Format Berkas:**
   - **Print / Save PDF:** Format cetak resmi dokumen A4.
   - **CSV Export:** Format spreadsheet tabel untuk Microsoft Excel.
   - **Download HTML:** Berkas salinan interaktif mandiri (*offline-ready*).
3. **Kolom Berkas CSV:** `Timestamp`, `Device Code`, `Parameter`, `Value`, `Unit`, `Quality Status`.

---

## Bab 11 — Pengaturan Pengguna dan Personalisasi

- **Pengaturan Bahasa:** Beralih antara Bahasa Indonesia dan English melalui tombol `EN | ID` di Header.
- **Tema Visual:** Mode Gelap (*Dark Mode*) untuk ruang kendali SCADA dan Mode Terang (*Light Mode*) untuk kantor.
- **Manajemen Akun (`/#/administration/users`):** Khusus Administrator customer untuk mendaftarkan akun operator baru, mengubah email/nama, dan mengunci akun non-aktif.

---

## Bab 12 — Panduan Troubleshooting Praktis Customer

| Gejala Masalah | Tindakan Mandiri yang Aman | Kapan Hubungi Vendor? |
| :--- | :--- | :--- |
| **Gagal Login** | Periksa tombol Caps Lock; ketik ulang kata sandi; minta Admin mereset sandi. | Jika semua akun admin terkunci. |
| **Halaman Web Tidak Dapat Dibuka** | Periksa kabel LAN PC dan power switch stasiun; lakukan uji `ping [IP_Server]`. | Jika power normal tetapi ping gagal total. |
| **Perangkat Sensor OFFLINE** | Periksa konektor kabel RS-485 pada terminal sensor; pastikan USB serial terpasang rapat. | Jika kabel normal namun offline > 30 menit. |
| **Nilai Sensor Membeku (Stuck Value)** | Periksa inlet sampling udara di luar stasiun; bersihkan filter debu jika tersumbat. | Jika pembersihan filter tidak mengubah nilai. |
| **Status Data UNCERTAIN / BAD** | Amati tren 15 menit; catat pada logbook; periksa apakah ada aktivitas kalibrasi. | Jika status BAD berlangsung > 1 jam. |
| **Grafik Tren Kosong** | Klik *Reset Filter*; pilih rentang waktu *Last 24 Hours*; centang parameter. | Tidak perlu. |
| **Layar HTTP 403 Forbidden** | Akun bertipe Operator mencoba membuka menu Admin; klik tombol *Operator Console*. | Tidak perlu (proteksi normal sistem). |
| **Ekspor Laporan Gagal** | Perkecil rentang waktu (misal per 7 hari atau per 30 hari); hindari rentang > 90 hari sekaligus. | Jika rentang 1 hari juga gagal. |
| **Jam Sistem Tidak Sinkron** | Periksa koneksi gateway internet di router stasiun agar sinkronisasi NTP berjalan. | Ya, teknisi akan menyinkronkan RTC. |
| **Aplikasi Terasa Lambat** | Tekan kombinasi tombol `Ctrl + F5` untuk membersihkan cache browser. | Tidak perlu. |

---

## Bab 13 — Prosedur Standar Operasional Harian (Daily SOP)

### 13.1 Rutinitas Awal Shift
- [ ] Buka browser dan login ke Operator Console.
- [ ] Verifikasi indikator stasiun berstatus **ONLINE** (Hijau).
- [ ] Pastikan denyut paket data (*Last packet*) < 15 detik.
- [ ] Periksa ikon lonceng header; pastikan tidak ada alarm *CRITICAL* aktif.

### 13.2 Pemantauan Berkala Shift
- [ ] Evaluasi nilai ISPU dan parameter kritis setiap 1–2 jam.
- [ ] Jika ada alarm berbunyi, lakukan *Acknowledge* dan catat pada logbook.
- [ ] Periksa kestabilan kurva trend multi-parameter bawah.

### 13.3 Rutinitas Akhir Shift
- [ ] Unduh rekapitulasi data shift pada menu Reports.
- [ ] Tulis catatan serah terima (*handover logbook*) untuk petugas shift berikutnya.
- [ ] Lakukan **Logout** akun secara aman sebelum meninggalkan ruang kendali.

---

## Bab 14 — Layanan Dukungan Teknis dan Eskalasi Vendor

### 14.1 Matriks Response Time (SLA)
- **Severity 1 (Kritis - Sistem Mati Total):** Respon < 1 Jam, Target Resolusi < 4 Jam.
- **Severity 2 (Tinggi - 1 Sensor Mati / Ekspor Error):** Respon < 2 Jam, Target Resolusi < 8 Jam kerja.
- **Severity 3 (Normal - Pertanyaan Operasional / Akun):** Respon < 4 Jam, Target Resolusi < 24 Jam kerja.

### 14.2 Kontak Resmi Layanan Dukungan
- **Helpdesk Portal:** `[https://support.namaperusahaan.co.id]`
- **Email Dukungan:** `[support@namaperusahaan.co.id]`
- **Hotline Telepon / WhatsApp Emergency:** `[0812-XXXX-XXXX / (021) XXXX-XXXX]`
- **Jam Operasional:** 24 Jam x 7 Hari (Insiden Kritis) / Jam Kerja 08.00 – 17.00 WIB (Pertanyaan Umum).

---

## Lampiran-Lampiran Lengkap

### Lampiran A: Glosarium Singkatan
- **AQMS:** Air Quality Monitoring System.
- **ISPU:** Indeks Standar Pencemar Udara (Permen LHK No. 14 Tahun 2020).
- **SCADA:** Supervisory Control and Data Acquisition.
- **Modbus RTU/TCP:** Protokol industri komunikasi sensor lapangan.
- **Data Completeness:** Persentase ketersediaan data valid dalam 24 jam ($\ge 75\%$).

### Lampiran B: Matriks Status Cepat
- **ONLINE (Hijau):** Polling komunikasi berhasil normal.
- **OFFLINE (Merah):** Timeout komunikasi > 2 detik.
- **GOOD (Hijau):** Data sensor valid dan akurat.
- **BAD (Merah):** Data di luar rentang atau terjadi spike.

### Lampiran C: Formulir Tiket Gangguan Lapangan
Saat melaporkan gangguan ke Helpdesk Vendor, sertakan:
1. Tanggal dan waktu kejadian persis (WIB).
2. Nama stasiun dan kode perangkat (misal: `AQMS-02`, Parameter: `pm2.5`).
3. Pesan error atau kode alarm persis (misal: `DATA_ABNORMALITY_STUCK_VALUE`).
4. Tangkapan layar (*screenshot*) halaman terkait.
5. Tindakan awal yang telah dicoba operator.
