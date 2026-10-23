# Datalogger Analysis Application — QA Test Procedure
**Document ID:** DAA-QA-001-PROC  
**Version:** 1.0  
**Status:** Draft for QA Execution / Release Candidate  
**Prepared by:** PT. Cipta Bina Instrumentasi (QA & Commissioning Team)  
**Project / Site:** AQMS Monitoring Station (Edge SCADA)  
**Target Platform:** Industrial Linux ARM64 / Windows 10 IoT x64  
**Date:** Oktober 2026  

---

## 1. Lembar Pengesahan Dokumen (Document Approval Matrix)

| Peran / Jabatan | Nama Lengkap | Organisasi / Departemen | Tanggal | Status |
| :--- | :--- | :--- | :---: | :---: |
| **QA Lead Engineer** (Penyusun) | [Nama QA Engineer] | PT. CBI - QA Dept | 2026-10-23 | Pending Execution |
| **Lead Instrumentation Engineer** (Pemeriksa Hardware) | [Nama Lead Instrumentation] | PT. CBI - Engineering | 2026-10-24 | Pending Witness |
| **System Analyst / Architect** (Pemeriksa Sistem & DB) | [Nama System Analyst] | PT. CBI - Software Dept | 2026-10-24 | Pending Review |
| **Project Manager / Commissioning Lead** (Otoritas Rilis) | [Nama Project Manager] | PT. CBI - Commissioning | 2026-10-25 | Pending Release |

### Riwayat Revisi Dokumen

| Versi | Tanggal | Penyusun | Ringkasan Perubahan | Status |
| :---: | :---: | :--- | :--- | :---: |
| **0.1** | 2026-10-20 | Tim QA CBI | Draf awal perancangan skenario 13 kategori lingkup uji | Draft |
| **0.9** | 2026-10-22 | QA & Eng CBI | Penyelarasan peta register sensor AQMS-01/02 dan SOP batch fallback | Review |
| **1.0** | 2026-10-23 | Lead QA CBI | Rilis resmi dokumen pendamping checklist XLSX DAA-QA-001 untuk eksekusi | Approved |

---

## 2. Pendahuluan & Ruang Lingkup Pengujian

### 2.1 Tujuan Dokumen
Dokumen Prosedur Pengujian Mutu (QA Test Procedure) ini disusun sebagai pedoman operasional baku bagi QA Engineer, System Analyst, Developer, Teknisi Instrumentasi, dan Tim Commissioning untuk memverifikasi keandalan, akurasi, dan keamanan aplikasi Datalogger Analysis Application sebelum dilakukan rilis dan serah terima (handover) ke pihak customer.

### 2.2 Ruang Lingkup Pengujian
Pengujian mencakup 13 kategori utama dengan total **162 concrete test cases**:
1. **Installation & Deployment (12 TCs):** Prasyarat sistem operasi, kompilasi binary Go, bundle frontend dist Vite, konektivitas MariaDB edge local, migrasi otomatis, lifecycle daemon, graceful shutdown, dan audit rahasia (secrets).
2. **Authentication & Authorization (12 TCs):** Autentikasi JWT, validasi password bcrypt, session expiry HTTP 401, route guards, RBAC (Operator, Engineer, Admin, Superadmin), dan credential logging hygiene.
3. **Device & Communication (14 TCs):** Manajemen perangkat Modbus RTU/TCP, konfigurasi baud rate 9600 8-N-1, slave ID validasi, timeout & retry policy, live test read register, status online/offline integrity, dan auto-reconnect serial.
4. **Parameter & Decoding (14 TCs):** Registrasi holding/input register, presisi decoding Float32 (ABCD vs CDAB word swap), INT16/32, kalkulasi linear scaling $y = mx + c$, formula kustom, dan penanganan nilai out-of-range / NaN.
5. **Polling Plan & Batch Read (14 TCs):** Optimasi batch read grouping, parameter max batch registers (32), max register gap (4), preview visual polling plan, pembuktian kesetaraan nilai batch vs single, dan mekanisme fallback otomatis saat batch failure.
6. **Telemetry & Data Integrity (12 TCs):** Siklus polling 5 detik, kualitas data (GOOD, BAD, UNCERTAIN), penanganan timestamp boundary timezone, pencegahan duplikasi data, dan integritas penyimpanan tabel `telemetry_readings`.
7. **Aggregation & Rollup (10 TCs):** Rollup otomatis 1 menit, 5 menit, 1 jam, dan 24 jam; verifikasi batas waktu stasiun; dan validasi matematis formula ISPU Permen LHK No. 14 Tahun 2020.
8. **Dashboard & UI (16 TCs):** Rendering kartu sensor SCADA, sparkline live 24 titik, filter time range, visualisasi wind rose & distribusi polutan, toggle tema light/dark mode, pemusatan modal dialog, dan responsivitas layout.
9. **Historical Data & Export (12 TCs):** Query rentang historis besar, pemisahan sumber raw vs aggregated, ekspor laporan berkala ke format CSV dan PDF, dan integritas format angka/satuan.
10. **Alarm, Logs & Audit Trail (12 TCs):** Evaluasi threshold warning/critical low & high, deteksi anomali stuck value (>= 3 siklus pembacaan identik), delay counter alarm, log audit aktivitas konfigurasi, dan rotasi file log.
11. **Backup, Restore & Retention (10 TCs):** Script mysqldump backup, kompresi gzip, verifikasi checksum file cadangan, simulasi pemulihan pada database terisolasi, dan kebijakan retensi pembersihan data lama (> 30 hari).
12. **Performance & Security (12 TCs):** Uji soak test 24 jam kontinu, pemantauan memori heap browser & RAM daemon, proteksi injeksi SQL, proteksi XSS pada nama sensor, dan penegakan prinsip least privilege user database.
13. **Customer Acceptance (12 TCs):** Uji terima lapangan (SAT) bersama customer, login mandiri operator customer, perbandingan pembacaan sensor dengan alat ukur kalibrasi, serah terima buku manual DAA-CUM-001, dan penandatanganan BAST.

---

## 3. Persiapan Lingkungan Uji (Test Environment Setup)

### 3.1 Topologi Testbed Staging
```
+-------------------------------------------------------------------------+
|                  HOST SERVER TESTBED (ARM64 / Windows 10)               |
|                                                                         |
|  +--------------------+   HTTP :8080    +----------------------------+  |
|  | Vue 2 SCADA Web    |<===============>| Go Backend Application     |  |
|  | (Vite Dist Bundle) |                 | (Datalogger Engine Daemon) |  |
|  +--------------------+                 +--------------+-------------+  |
|                                                        |                |
|  +--------------------+                                | Local Unix /   |
|  | MariaDB 10.4+ Local|<================---------------+ TCP Port 3306  |
|  | Database 'datalogger'                               |                |
|  +--------------------+                                |                |
+--------------------------------------------------------|----------------+
                                                         |
                   +-------------------------------------+
                   | Serial /dev/ttyUSB0 (COM3)
                   | Baud: 9600, 8-N-1
                   v
       +-----------------------+
       | Isolated USB-to-RS485 |
       | Industrial Converter  |
       +-----------+-----------+
                   |
     ==============+== Bus RS-485 (Shielded Twisted Pair, 120 Ohm) ============
                   |                                       |
                   v                                       v
       +-----------------------+               +-----------------------+
       | AQMS-01: METEOROLOGI  |               | AQMS-02: AIR QUALITY  |
       | Modbus Slave ID: 1    |               | Modbus Slave ID: 2 / 6|
       | Suhu, Kelembaban,     |               | PM2.5, PM10, SO2,     |
       | Tekanan, Angin, Hujan |               | NO2, CO, O3, HC       |
       +-----------------------+               +-----------------------+
```

### 3.2 Prasyarat Perangkat Keras
- **Host Edge:** Industrial Box PC / Raspberry Pi 4B (ARM64) atau PC Intel Core i3 x64.
- **RAM:** Minimal 2 GB (Rekomendasi 4 GB).
- **Storage:** Minimal 16 GB eMMC / High-Endurance SSD.
- **Serial Converter:** USB to RS-485 industri dengan isolasi optocoupler (Galvanic Isolation 2.5 kV) berbasis chip FTDI atau Silicon Labs CP2102.
- **Kabel Bus:** Twisted Pair Shielded (Belden 9841 atau setara) dengan resistor terminasi 120 $\Omega$ pada kedua ujung bus.

### 3.3 Setup Database Edge Lokal
```bash
# Inisialisasi user dan database
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS datalogger CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
mysql -u root -p -e "CREATE USER IF NOT EXISTS 'datalogger'@'127.0.0.1' IDENTIFIED BY 'DataloggerSecret2026!';"
mysql -u root -p -e "GRANT ALL PRIVILEGES ON datalogger.* TO 'datalogger'@'127.0.0.1'; FLUSH PRIVILEGES;"

# Jalankan daemon backend untuk migrasi otomatis skema
./datalogger
```

---

## 4. Ground-Truth Test Datasets & Peta Register

### 4.1 Peta Register Sensor AQMS-01 (Meteorologi)
- **Slave ID:** 1 | **Baud:** 9600 8-N-1 | **Format:** Float32 Word Swap (CDAB)

| Parameter | Alamat Hex | Alamat Des | Tipe Register | Format Data | Satuan | Rentang Valid |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: |
| Suhu Lingkungan (Temp) | 0x0000 | 0 | Holding | Float32 CDAB | °C | -20.0 ~ 70.0 |
| Kelembaban Relatif (Hum) | 0x0002 | 2 | Holding | Float32 CDAB | %RH | 0.0 ~ 100.0 |
| Tekanan Udara (Pressure) | 0x0004 | 4 | Holding | Float32 CDAB | hPa | 500 ~ 1200 |
| Kecepatan Angin (Wind Speed)| 0x0006 | 6 | Holding | Float32 CDAB | m/s | 0.0 ~ 60.0 |
| Arah Angin (Wind Direction) | 0x0008 | 8 | Holding | Float32 CDAB | deg (°) | 0.0 ~ 360.0 |
| Radiasi Surya (Solar Rad)  | 0x000A | 10 | Holding | Float32 CDAB | W/m² | 0.0 ~ 1500 |
| Curah Hujan (Rainfall)      | 0x000C | 12 | Holding | Float32 CDAB | mm | 0.0 ~ 500.0 |

### 4.2 Peta Register Sensor AQMS-02 (Kualitas Udara & Gas)
- **Slave ID:** 6 (atau 2) | **Baud:** 9600 8-N-1 | **Format:** Float32 Word Swap (CDAB)

| Parameter | Alamat Hex | Alamat Des | Tipe Register | Format Data | Satuan | Rentang Valid |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: |
| Partikulat PM2.5 | 0x0010 | 16 | Holding | Float32 CDAB | µg/m³ | 0.0 ~ 1000.0 |
| Partikulat PM10  | 0x0012 | 18 | Holding | Float32 CDAB | µg/m³ | 0.0 ~ 2000.0 |
| Sulfur Dioksida (SO2) | 0x0014 | 20 | Holding | Float32 CDAB | µg/m³ | 0.0 ~ 1500.0 |
| Nitrogen Dioksida (NO2)| 0x0016 | 22 | Holding | Float32 CDAB | µg/m³ | 0.0 ~ 1200.0 |
| Karbon Monoksida (CO) | 0x0018 | 24 | Holding | Float32 CDAB | µg/m³ | 0.0 ~ 50000 |
| Ozon (O3)             | 0x001A | 26 | Holding | Float32 CDAB | µg/m³ | 0.0 ~ 1000.0 |
| Hidrokarbon (HC)      | 0x001C | 28 | Holding | Float32 CDAB | µg/m³ | 0.0 ~ 1000.0 |

### 4.3 Tabel Baku Nilai Uji Heksadesimal IEEE-754

| Nilai Referensi Teknik | IEEE-754 (ABCD) | Word-Swapped (CDAB) | Reg 1 Hex | Reg 2 Hex |
| :--- | :---: | :---: | :---: | :---: |
| **25.40 °C** (Suhu) | 0x41CB3333 | 0x333341CB | `0x3333` | `0x41CB` |
| **65.50 %RH** (Kelembaban) | 0x42830000 | 0x00004283 | `0x0000` | `0x4283` |
| **1013.25 hPa** (Tekanan) | 0x447D5000 | 0x5000447D | `0x5000` | `0x447D` |
| **35.80 µg/m³** (PM2.5) | 0x420F3333 | 0x3333420F | `0x3333` | `0x420F` |
| **150.00 µg/m³** (PM10) | 0x43160000 | 0x00004316 | `0x0000` | `0x4316` |
| **2500.0 µg/m³** (CO) | 0x451C4000 | 0x4000451C | `0x4000` | `0x451C` |

---

## 5. Prosedur Pengujian Kritis Langkah-demi-Langkah (Critical SOPs)

### SOP-01: Uji Batch Read Optimization, Maximum Gap & Fallback
1. Login sebagai pengguna dengan hak akses konfigurasi (`ENGINEER` atau `SUPERADMIN`).
2. Masuk ke menu **Monitoring -> Devices**, pilih perangkat **AQMS-01**, lalu klik **Edit Device**.
3. Pada tab **Communication**, ubah *Batch Read Mode* ke `AUTO`, tetapkan *Max Batch Registers* = `32` dan *Max Register Gap* = `4`. Simpan konfigurasi.
4. Buka tab **Polling Plan Preview**. Pastikan sistem menampilkan 1 query Modbus holding register (Address: 0, Count: 14) alih-alih 7 query individual.
5. Verifikasi bahwa nilai seluruh parameter di Operator Console identik dengan mode Single Read.
6. Simulasikan kegagalan: Pada simulator Modbus, atur register 0x0006 menjadi Illegal Address (Exception 02).
7. Amati perilaku sistem: Batch request gagal, circuit breaker aktif, sistem otomatis fallback ke Single Read per-parameter. Parameter sehat lainnya tetap terbaca normal tanpa interupsi daemon.

### SOP-02: Uji Fault Injection Komunikasi Serial & Auto-Recovery
1. Pastikan status perangkat **ONLINE** (indikator hijau) dan telemetri terbarukan secara berkala.
2. Lepaskan kabel komunikasi fisik RS-485 (jalur A+ atau B-) secara mendadak.
3. Amati log backend: Terjadi 3x percobaan ulang (retry) sesuai konfigurasi timeout.
4. Setelah percobaan gagal terpenuhi, verifikasi di Operator Console bahwa status perangkat bertransisi ke **OFFLINE** (merah), alarm `COMM_MODBUS_TIMEOUT` aktif, dan bendera kualitas telemetri diset ke `BAD`.
5. Sambungkan kembali kabel komunikasi RS-485.
6. Amati proses auto-recovery: Dalam siklus polling berikutnya (< 10 detik), koneksi pulih secara mandiri, status kembali **ONLINE**, dan alarm otomatis ter-resolve.

### SOP-03: Uji Telemetri Real-Time & Verifikasi Kadensi 5 Detik
1. Login sebagai `OPERATOR` dan buka **Operator Console** pada Google Chrome.
2. Buka DevTools (F12) tab **Performance** dan **Network**.
3. Verifikasi request telemetri terjadi setiap 5000ms ± 200ms secara konsisten.
4. Amati sparkline 24 titik: Titik bergerak ke kiri dengan halus saat data baru diterima tanpa layout reflow atau kedipan layar.
5. Jalankan selama 30 menit dan pantau memori heap browser: Grafik memory harus stabil mendatar tanpa tanda kebocoran memori.

### SOP-04: Uji Agregasi Periodik & Formula ISPU Permen LHK 14/2020
1. Siapkan konsentrasi rata-rata 24 jam untuk 7 parameter polutan:
   - PM2.5 = 45.0 µg/m³
   - PM10 = 85.0 µg/m³
   - SO2 = 30.0 µg/m³
   - NO2 = 40.0 µg/m³
   - CO = 2500.0 µg/m³
   - O3 = 60.0 µg/m³
   - HC = 15.0 µg/m³
2. Hitung manual nilai ISPU PM2.5 berdasarkan tabel batas ambang (15.6 - 55.4 µg/m³ -> ISPU 51 - 100):
   $$I = \frac{100 - 51}{55.4 - 15.6} \times (45.0 - 15.6) + 51 = \frac{49}{39.8} \times 29.4 + 51 = 36.19 + 51 = 87.19 \approx 87$$
3. Buka modul **ISPU Analysis** di aplikasi web.
4. Verifikasi bahwa nilai ISPU sistem tepat bernilai **87**, masuk kategori **SEDANG (Biru)**, dan PM2.5 ditetapkan sebagai parameter kritis stasiun karena bernilai ISPU tertinggi.

### SOP-05: Uji Engine Alarm & Deteksi Anomali Stuck Value
1. Tetapkan threshold warning PM2.5 = 55.0 µg/m³ dengan delay counter = 2 siklus.
2. Berikan nilai 65.0 µg/m³. Verifikasi alarm terpicu pada siklus ke-2 dengan indikator lonceng merah berdering.
3. Suntikkan nilai sensor beku (stuck reading): Masukkan nilai konstan tepat 35.80 selama 4 siklus beruntun.
4. Engine mendeteksi variansi nol, memicu alarm `DATA_ABNORMALITY_STUCK_VALUE` untuk memperingatkan teknisi kemungkinan sensor membeku/hang.
5. Kembalikan ke nilai bervariasi: Alarm stuck value otomatis terselesaikan (*auto-cleared*).

### SOP-06: Uji Backup Database Fisik & Restorasi Terisolasi
1. Jalankan backup database: `mysqldump -u datalogger -p datalogger | gzip > backup_test.sql.gz`.
2. Pastikan file backup valid, tidak nol byte, dan berukuran > 50 KB.
3. Buat database uji terisolasi: `CREATE DATABASE datalogger_restore_test;`.
4. Lakukan pemulihan: `gunzip < backup_test.sql.gz | mysql -u datalogger -p datalogger_restore_test`.
5. Bandingkan jumlah baris tabel `telemetry_readings` dan `parameters` antara database sumber dan hasil restorasi: Seluruh data harus identik 100%.

---

## 6. Manajemen Cacat & Aturan Gerbang Rilis (Release Gate)

### 6.1 Matriks Keparahan Cacat (Severity Matrix)
- **CRITICAL (P1):** Kebocoran keamanan serius, crash panic daemon, kehilangan data telemetri, atau formula ISPU salah. Harus selesai dalam waktu < 12 jam.
- **HIGH (P2):** Fungsi utama gagal dan tidak ada workaround (misal: tombol ekspor PDF gagal, batch read tidak fallback). Harus selesai dalam waktu < 24 jam.
- **MEDIUM (P3):** Sebagian fungsi terganggu tetapi tersedia workaround yang dapat diterima (misal: query historis rentang panjang lambat). Harus selesai dalam kurun 3 hari kerja.
- **LOW (P4):** Isu minor estetika atau typo pada teks petunjuk.

### 6.2 Kriteria Kelulusan Rilis (Release Gate Criteria)
Aplikasi **DILARANG DINYATAKAN RILIS** jika belum memenuhi ambang batas berikut:
1. **100% Test Case Dieksekusi:** Seluruh 162 test case telah dieksekusi (0 kasus berstatus `NOT TESTED`).
2. **Nol Cacat Kritis & Tinggi:** 0 Open Critical Defects dan 0 Open High Defects pada sheet *Defect Log*.
3. **Validasi Hardware Fisik:** 100% pengujian komunikasi fisik pada bus RS-485 lulus tanpa frame error.
4. **Uji Ketahanan 24 Jam Nonstop:** Daemon berjalan kontinu selama 24 jam dengan interval polling 5 detik tanpa crash atau kebocoran memori.
5. **Dokumentasi Lengkap:** Buku Manual Pengguna (DAA-CUM-001) telah diserahkan dan Berita Acara Uji Terima (SAT) ditandatangani bersama customer.

---

## 7. Inventaris Test Case yang Wajib Menggunakan Perangkat Keras Fisik

| Test Case ID | Nama Pengujian | Perangkat Keras yang Diwajibkan |
| :---: | :--- | :--- |
| **TC-DEV-005** | Konfigurasi Serial Port & Baud Rate 9600 8-N-1 | Bus RS-485 fisik, converter USB-Serial |
| **TC-DEV-010** | Deteksi Otomatis OFFLINE saat Kabel Terputus | Terminal kabel RS-485 dilepas |
| **TC-DEV-011** | Penanganan Serial Port Busy Resource Lock | Dua aplikasi mengakses port serial yang sama |
| **TC-DEV-012** | Auto-Reconnect setelah Converter USB Dicabut | Adapter USB-RS485 dicabut & dipasang ulang |
| **TC-COMM-001**| Pembacaan Modbus RTU pada Sensor Riil | Instrumen sensor AQMS-01 & AQMS-02 di rak |
| **TC-COMM-009**| Pemutusan Fisik Sinyal Komunikasi A+/B- | Saklar pemutus jalur komunikasi RS-485 |
| **TC-COMM-012**| Pelepasan Converter saat Siklus Polling Berjalan | Kabel USB ditarik saat transmisi aktif |
| **TC-PERF-008**| Injeksi Gangguan Jaringan Fisik & Packet Loss | Managed Network Switch dengan traffic control |
| **TC-COMP-001**| Deployment pada Arsitektur ARM64 Fisik | Industrial Box PC Linux ARM64 |
| **TC-ACC-004** | Uji Banding Sensor dengan Alat Ukur Standar | Gas kalibrasi bersertifikat / alat ukur acuan |

---

*Dokumen ini merupakan bagian tak terpisahkan dari Berita Acara Serah Terima (BAST) sistem Datalogger Analysis Application.*
