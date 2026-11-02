# Panduan Instalasi Multi-Platform Datalogger Analysis Application (via GitHub Clone)

Dokumen ini merupakan panduan resmi instalasi, konfigurasi, pembaruan (*update*), dan pemeliharaan **Datalogger Analysis Application** pada berbagai sistem operasi dan arsitektur perangkat keras melalui repositori GitHub.

---

## Daftar Isi

1. [Arsitektur & Kebutuhan Sistem](#1-arsitektur--kebutuhan-sistem)
2. [Persiapan Awal: Clone Repositori GitHub](#2-persiapan-awal-clone-repositori-github)
3. [Instalasi pada Linux & Raspberry Pi (Ubuntu / Debian / Raspberry Pi OS)](#3-instalasi-pada-linux--raspberry-pi-ubuntu--debian--raspberry-pi-os)
4. [Instalasi pada Windows (Windows 10 / 11 / Server / Industrial PC)](#4-instalasi-pada-windows-windows-10--11--server--industrial-pc)
5. [Instalasi pada macOS (Intel & Apple Silicon M1–M4)](#5-instalasi-pada-macos-intel--apple-silicon-m1m4)
6. [Konfigurasi Aplikasi (`config.json` & Environment Variables)](#6-konfigurasi-aplikasi-configjson--environment-variables)
7. [Setup Port USB-to-RS485 / Modbus Serial](#7-setup-port-usb-to-rs485--modbus-serial)
8. [Akses Dashboard & Verifikasi Pasca-Instalasi](#8-akses-dashboard--verifikasi-pasca-instalasi)
9. [Prosedur Update Aplikasi (via `git pull`)](#9-prosedur-update-aplikasi-via-git-pull)
10. [Cross-Compilation (Build dari 1 Mesin untuk Semua Platform)](#10-cross-compilation-build-dari-1-mesin-untuk-semua-platform)
11. [Uninstall & Troubleshooting](#11-uninstall--troubleshooting)

---

## 1. Arsitektur & Kebutuhan Sistem

Aplikasi dirancang dengan arsitektur **Local-First Edge** menggunakan **Pure-Go (`CGO_ENABLED=0`)** sehingga tidak membutuhkan *C compiler* eksternal (`gcc`) saat kompilasi. Seluruh file *production frontend* (`web/dist`) sudah disertakan langsung di dalam repositori, sehingga mesin target **tidak wajib** menginstal Node.js kecuali Anda ingin memodifikasi kode sumber tampilan Web UI.

### Matriks Platform yang Didukung

| Platform | Arsitektur CPU | Mode Service Otomatis | Direktori Instalasi Default |
| :--- | :--- | :--- | :--- |
| **Linux Server / Mini PC** | `x86_64` (`amd64`) | `systemd` (`datalogger.service`) | `/opt/datalogger` |
| **Raspberry Pi 4 / 5** | `aarch64` (`arm64`) | `systemd` (`datalogger.service`) | `/opt/datalogger` |
| **Raspberry Pi 2 / 3 / Zero 2** | `armv7l` / `armhf` | `systemd` (`datalogger.service`) | `/opt/datalogger` |
| **Windows 10 / 11 / Server** | `x64` (`amd64`) | Startup Folder / Task Scheduler / NSSM | Direktori hasil `git clone` |
| **macOS** | Apple Silicon (`arm64`) & Intel (`x86_64`) | `launchd` (`com.datalogger.analysis`) | `/usr/local/opt/datalogger` |

### Spesifikasi Perangkat Keras Minimum & Rekomendasi

- **CPU:** Minimal Dual-Core 1.0 GHz (Rekomendasi: Quad-Core ARM Cortex-A72 / Intel Celeron / Core i3 ke atas)
- **RAM:** Minimal 1 GB (Rekomendasi: 2 GB – 4 GB+)
- **Penyimpanan:** Minimal ruang kosong 2 GB (Rekomendasi: SSD / Industrial MicroSD 16 GB+ untuk penyimpanan historis telemetri & *backup*)
- **Perangkat Lunak Prasyarat:**
  1. **Git** (`2.x+`)
  2. **Go (Golang)** versi `1.22` ke atas (disarankan `1.26+` jika melakukan kompilasi langsung dari kode sumber)
  3. **MariaDB Server** (`10.5+` / `10.11+` / `11.x`) atau MySQL (`8.0+`) yang berjalan di `127.0.0.1:3306`
  4. *(Opsional)* **Node.js `18+` & npm** — hanya diperlukan apabila ingin melakukan *rebuild* ulang folder `web/`

---

## 2. Persiapan Awal: Clone Repositori GitHub

Jalankan perintah berikut pada terminal (Bash / PowerShell / Terminal macOS) di mesin target:

```bash
git clone https://github.com/acepahmads/datalogger-27.git
cd datalogger-27
```

> **Catatan Repository Private:**
> Jika repositori bersifat *private*, gunakan Personal Access Token (PAT) atau SSH Key:
> ```bash
> # Opsi HTTPS dengan Token:
> git clone https://<GITHUB_TOKEN>@github.com/acepahmads/datalogger-27.git
>
> # Opsi SSH:
> git clone git@github.com:acepahmads/datalogger-27.git
> ```

---

## 3. Instalasi pada Linux & Raspberry Pi (Ubuntu / Debian / Raspberry Pi OS)

Platform Linux dan Raspberry Pi dilengkapi skrip instalasi otomatis penuh (`scripts/install-linux.sh`) yang menangani instalasi MariaDB, pembuatan database & user, kompilasi *binary* sesuai arsitektur CPU, penyalinan aset Web UI, hingga registrasi *daemon* `systemd`.

### 3.1. Instalasi Prasyarat Dasar (Jika Belum Ada)

```bash
sudo apt-get update
sudo apt-get install -y git curl mariadb-server
```

> **Catatan Versi Go di Debian/Ubuntu/Raspberry Pi:**
> Skrip installer akan otomatis mencoba menginstal `golang-go` jika belum tersedia. Namun jika repositori bawaan OS memiliki versi Go yang terlalu lama (`< 1.22`), instal Go resmi terbaru terlebih dahulu:
> ```bash
> # Contoh untuk Linux x86_64 (ganti 'amd64' menjadi 'arm64' untuk Raspberry Pi OS 64-bit, atau 'armv6l' untuk 32-bit):
> ARCH_GO=$(dpkg --print-architecture) # menghasilkan amd64, arm64, atau armhf
> [ "$ARCH_GO" = "armhf" ] && ARCH_GO="armv6l"
> curl -LO "https://go.dev/dl/go1.26.4.linux-${ARCH_GO}.tar.gz"
> sudo rm -rf /usr/local/go
> sudo tar -C /usr/local -xzf "go1.26.4.linux-${ARCH_GO}.tar.gz"
> export PATH=$PATH:/usr/local/go/bin
> echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
> go version
> ```

### 3.2. Menjalankan Installer Otomatis (`systemd`)

Dari dalam folder hasil clone (`datalogger-27`), jalankan:

```bash
sudo env "PATH=$PATH:/usr/local/go/bin" bash ./scripts/install-linux.sh
```

**Apa yang dilakukan otomatis oleh `install-linux.sh`:**
1. Membuat direktori standar industri:
   - Aplikasi & Web UI: `/opt/datalogger`
   - Data & Queue Spool: `/var/lib/datalogger`
   - Log Sistem: `/var/log/datalogger`
2. Memastikan `mariadb.service` aktif, membuat database `datalogger` (`utf8mb4`), serta membuat akun database lokal `datalogger` / `datalogger`.
3. Menyalin binary pra-kompilasi (jika tersedia di `bin/`) atau mengompilasi langsung `cmd/main.go` ke `/opt/datalogger/datalogger`.
4. Menyalin `web/dist` ke `/opt/datalogger/web/dist` dan `config.json` ke `/opt/datalogger/config.json` (tanpa menimpa `config.json` lama saat *upgrade*).
5. Mendaftarkan dan menjalankan service `/etc/systemd/system/datalogger.service` dengan `Restart=always` (otomatis menyala kembali saat *reboot* atau mati listrik).

### 3.3. Perintah Manajemen Service Linux (`systemd`)

```bash
# Cek status service
sudo systemctl status datalogger.service

# Melihat log realtime aplikasi
sudo journalctl -u datalogger.service -f

# Restart / Stop / Start service
sudo systemctl restart datalogger.service
sudo systemctl stop datalogger.service
sudo systemctl start datalogger.service
```

---

## 4. Instalasi pada Windows (Windows 10 / 11 / Server / Industrial PC)

### 4.1. Instalasi Prasyarat Windows

1. **Git for Windows:** Unduh & instal dari https://git-scm.com/download/win
2. **Go (Golang) 64-bit:** Unduh & instal MSI dari https://go.dev/dl/
3. **MariaDB Server Windows:**
   - Unduh MSI Installer dari https://mariadb.org/download/
   - Saat instalasi, centang **Install as service** (port default `3306`) dan catat *root password* Anda (atau kosongkan jika menggunakan jaringan lokal tertutup sesuai kebijakan keamanan).

### 4.2. Konfigurasi Kredensial Database Awal

Buka PowerShell, clone repositori, lalu sesuaikan `config.json` dengan user dan password MariaDB Windows Anda:

```powershell
git clone https://github.com/acepahmads/datalogger-27.git
cd datalogger-27
```

Periksa bagian database di `config.json`:

```json
{
  "port": "8080",
  "host": "0.0.0.0",
  "db_type": "mariadb",
  "db_host": "127.0.0.1",
  "db_port": "3306",
  "db_name": "datalogger",
  "db_user": "root",
  "db_password": "PASSWORD_MARIADB_ANDA"
}
```
*(Catatan: Database `datalogger` serta seluruh tabel akan dibuat otomatis oleh aplikasi saat pertama kali dijalankan).*

### 4.3. Opsi A — Instalasi Otomatis via PowerShell Installer (Direkomendasikan untuk Industrial PC)

Jalankan skrip `install-windows.ps1` melalui PowerShell:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\install-windows.ps1
```

**Apa yang dilakukan oleh `install-windows.ps1`:**
1. Menyiapkan folder runtime `data\`, `logs\`, dan `bin\`.
2. Memverifikasi aset `web\dist\index.html` (otomatis menjalankan `npm run build` jika belum ada).
3. Mengompilasi binary murni `bin\datalogger.exe` (`CGO_ENABLED=0`).
4. Membuat peluncur `start-datalogger.bat` dan mendaftarkan *shortcut* `Datalogger-AutoStart.lnk` di Windows Startup Folder agar aplikasi otomatis berjalan saat PC menyala ulang.
5. Menjalankan aplikasi dan membuka browser ke `http://localhost:8080`.

> **Opsi Parameter Tambahan PowerShell Installer:**
> - Instal tanpa langsung menjalankan aplikasi:
>   `powershell -ExecutionPolicy Bypass -File .\scripts\install-windows.ps1 -NoLaunch`
> - Instal tanpa mendaftarkan Auto-Start:
>   `powershell -ExecutionPolicy Bypass -File .\scripts\install-windows.ps1 -AutoStart:$false`

### 4.4. Opsi B — Menjalankan sebagai Windows Background Service (via NSSM untuk Windows Server)

Jika PC beroperasi tanpa *user login* (*headless server*), Anda dapat mendaftarkan `bin\datalogger.exe` sebagai Windows Service menggunakan **NSSM (Non-Sucking Service Manager)**:

```powershell
# Setelah menjalankan build:
$env:CGO_ENABLED = "0"
go build -ldflags "-s -w" -o .\bin\datalogger.exe .\cmd\main.go

# Daftarkan ke Windows Service dengan NSSM:
nssm install DataloggerService "$PWD\bin\datalogger.exe"
nssm set DataloggerService AppDirectory "$PWD"
nssm set DataloggerService AppStdout "$PWD\logs\service-stdout.log"
nssm set DataloggerService AppStderr "$PWD\logs\service-stderr.log"
nssm start DataloggerService
```

---

## 5. Instalasi pada macOS (Intel & Apple Silicon M1–M4)

### 5.1. Instalasi Prasyarat via Homebrew

```bash
# Instal Git, Go, dan MariaDB menggunakan Homebrew
brew install git go mariadb

# Jalankan MariaDB sebagai background service
brew services start mariadb

# Buat akun database 'datalogger' (atau sesuaikan config.json)
sudo mariadb -e "CREATE DATABASE IF NOT EXISTS datalogger CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
sudo mariadb -e "CREATE USER IF NOT EXISTS 'datalogger'@'127.0.0.1' IDENTIFIED BY 'datalogger';"
sudo mariadb -e "GRANT ALL PRIVILEGES ON datalogger.* TO 'datalogger'@'127.0.0.1'; FLUSH PRIVILEGES;"
```

### 5.2. Clone & Jalankan Installer `launchd` macOS

```bash
git clone https://github.com/acepahmads/datalogger-27.git
cd datalogger-27

# Sesuaikan db_user & db_password pada config.json bila perlu, lalu jalankan installer:
sudo bash ./scripts/install-darwin.sh
```

**Apa yang dilakukan oleh `install-darwin.sh`:**
- Mendeteksi arsitektur Mac (`arm64` Apple Silicon atau `x86_64` Intel).
- Mengompilasi binary ke `/usr/local/opt/datalogger/datalogger` dan menyalin `web/dist` serta `config.json`.
- Memasang service `launchd` di `/Library/LaunchDaemons/com.datalogger.analysis.plist` dengan `RunAtLoad` & `KeepAlive` sehingga otomatis aktif saat booting.
- Log dapat dipantau di `/var/log/datalogger/stdout.log` dan `/var/log/datalogger/stderr.log`.

---

## 6. Konfigurasi Aplikasi (`config.json` & Environment Variables)

Aplikasi membaca konfigurasi dengan urutan prioritas:
1. **Environment Variables** (prioritas tertinggi — menimpa nilai di `config.json`)
2. **File `config.json`** (dicari di `DATALOGGER_CONFIG`, direktori kerja saat ini, atau berdampingan dengan file binary)
3. **Nilai Default Internal**

### Tabel Referensi Parameter Penting

| Kunci `config.json` | Environment Variable | Default | Deskripsi |
| :--- | :--- | :--- | :--- |
| `port` | `DATALOGGER_PORT` | `8080` | Port HTTP Web UI, REST API, & WebSocket |
| `host` | `DATALOGGER_HOST` | `0.0.0.0` | Bind IP (`0.0.0.0` agar dapat diakses dari LAN) |
| `db_type` | `DB_DRIVER` / `DATALOGGER_DB_TYPE` | `mariadb` | Engine database (`mariadb` / `mysql`) |
| `db_host` | `DB_HOST` / `DATALOGGER_DB_HOST` | `127.0.0.1` | Host server MariaDB |
| `db_port` | `DB_PORT` / `DATALOGGER_DB_PORT` | `3306` | Port TCP MariaDB |
| `db_name` | `DB_NAME` / `DATALOGGER_DB_NAME` | `datalogger` | Nama skema database |
| `db_user` | `DB_USER` / `DATALOGGER_DB_USER` | `root` | Username database |
| `db_password` | `DB_PASSWORD` / `DATALOGGER_DB_PASSWORD` | `""` | Password database |
| `jwt_secret` | `DATALOGGER_JWT_SECRET` | *(bawaan)* | Kunci rahasia token sesi login JWT (wajib diganti di produksi publik) |
| `data_dir` | `DATALOGGER_DATA_DIR` | `data` | Direktori penyimpanan runtime, queue WAL, & backup |
| `log_dir` | `DATALOGGER_LOG_DIR` | `logs` | Direktori log aplikasi terstruktur |
| *(N/A)* | `DATALOGGER_WEB_DIR` | `./web/dist` | Path eksplisit ke folder statis frontend Vue |
| `monitoring_domain` | `DATALOGGER_MONITORING_DOMAIN` | `AQMS` | Mode domain utama (`AQMS`, `WQMS`, atau `CUSTOM`) |

### Rebuild Frontend Web UI (Opsional — Hanya Jika Mengubah Kode `web/src`)

Jika Anda memodifikasi file Vue di dalam `web/src/`, jalankan perintah berikut sebelum menjalankan ulang aplikasi:

```bash
cd web
npm install
npm run build
cd ..
```

---

## 7. Setup Port USB-to-RS485 / Modbus Serial

### 7.1. Linux & Raspberry Pi (Mengunci Port USB agar Tidak Berubah `ttyUSB0` ↔ `ttyUSB1`)

Pada Linux/Raspberry Pi, nama port `/dev/ttyUSB0` dan `/dev/ttyUSB1` dapat tertukar setelah *reboot* atau saat menggunakan lebih dari satu konverter USB-to-RS485. Tancapkan seluruh konverter USB-RS485 Anda, lalu jalankan skrip bawaan:

```bash
sudo bash ./scripts/setup-usb-udev.sh
```

Skrip ini akan:
- Mendeteksi *Chip Serial Number* (FTDI / CP210x) atau *Physical USB Socket Path* (untuk chip CH340 tanpa serial unik).
- Membuat aturan permanen di `/etc/udev/rules.d/99-datalogger-serial.rules` dengan alias tetap seperti:
  - `/dev/datalogger_rs485_1`
  - `/dev/datalogger_rs485_2`
- Gunakan path `/dev/datalogger_rs485_1` tersebut saat mendaftarkan perangkat Modbus RTU di menu **Devices** pada Web UI.

*(Jika menjalankan aplikasi sebagai user non-root di Linux, pastikan user tergabung ke grup `dialout`: `sudo usermod -aG dialout $USER` lalu relogin).*

### 7.2. Windows

1. Buka **Device Manager → Ports (COM & LPT)** untuk melihat nomor port konverter USB-to-RS485 (misalnya `COM3`, `COM4`).
2. Untuk mencegah nomor COM berubah saat berpindah slot USB, klik kanan pada port di Device Manager → **Properties → Port Settings → Advanced**, lalu kunci **COM Port Number** ke nomor tetap (misalnya `COM10`, `COM11`).

---

## 8. Akses Dashboard & Verifikasi Pasca-Instalasi

1. Buka browser (Chrome / Edge / Firefox) dan akses:
   - Dari mesin yang sama: `http://localhost:8080`
   - Dari jaringan LAN yang sama: `http://<IP_ADDRESS_MESIN>:8080`
2. **Akun Login Default Pertama Kali:**
   - **Username:** `admin`
   - **Password:** `admin123`
   *(Segera ubah password default melalui menu manajemen pengguna / pengaturan setelah login pertama).*
3. **Checklist Verifikasi Cepat:**
   - Buka menu **System Health / Activity Monitor** → pastikan indikator **MariaDB Status = Connected / Healthy** dan **WebSocket = Connected**.
   - Jika diakses dari komputer lain di jaringan LAN dan halaman tidak terbuka, izinkan port `8080` pada Firewall:
     - **Linux (UFW):** `sudo ufw allow 8080/tcp`
     - **Windows (PowerShell Admin):**
       `New-NetFirewallRule -DisplayName "Datalogger Web UI" -Direction Inbound -Protocol TCP -LocalPort 8080 -Action Allow`

---

## 9. Prosedur Update Aplikasi (via `git pull`)

Seluruh pembaruan skema tabel database berjalan otomatis (*Auto-Migration*) saat binary terbaru dinyalakan tanpa menghapus data telemetri maupun konfigurasi yang sudah ada.

### 9.1. Update 1-Baris pada Linux & Raspberry Pi

Masuk ke folder repositori hasil clone, lalu jalankan skrip `update-linux.sh`:

```bash
cd ~/datalogger-27
sudo env "PATH=$PATH:/usr/local/go/bin" bash ./scripts/update-linux.sh
```
*(Skrip akan melakukan `git pull origin main`, menghentikan sementara `datalogger.service`, mengompilasi/menyalin binary & `web/dist` terbaru ke `/opt/datalogger`, lalu menyalakan kembali service).*

### 9.2. Update pada Windows

```powershell
cd D:\path\ke\datalogger-27

# 1. Hentikan proses datalogger yang sedang berjalan
Stop-Process -Name "datalogger" -Force -ErrorAction SilentlyContinue

# 2. Tarik kode & aset UI terbaru dari GitHub
git pull origin main

# 3. Kompilasi ulang binary terbaru
$env:CGO_ENABLED = "0"
go build -ldflags "-s -w" -o .\bin\datalogger.exe .\cmd\main.go

# 4. Jalankan kembali aplikasi
.\start-datalogger.bat
```

### 9.3. Update pada macOS

```bash
cd ~/datalogger-27
git pull origin main
sudo bash ./scripts/install-darwin.sh
```

> **Tips Browser Setelah Update:** Tekan `Ctrl + Shift + R` (Windows/Linux) atau `Cmd + Shift + R` (macOS) untuk melakukan *Hard Refresh* agar aset antarmuka terbaru langsung dimuat.

---

## 10. Cross-Compilation (Build dari 1 Mesin untuk Semua Platform)

Karena aplikasi menggunakan `CGO_ENABLED=0`, Anda dapat mengompilasi *binary* untuk seluruh OS/arsitektur langsung dari **satu komputer** (misalnya laptop Windows atau Linux), lalu menyalinnya ke folder `bin/` sebelum di-deploy ke perangkat *edge* spesifikasi rendah (seperti Raspberry Pi Zero/3 yang lambat jika harus mengompilasi sendiri):

### Dari Windows (Otomatis Semua Target)

```bat
.\scripts\build-cross-platform.bat
```

### Perintah Manual per Target (Bash Linux / macOS)

```bash
mkdir -p bin

# 1. Windows 64-bit (x64)
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o bin/datalogger-windows-amd64.exe ./cmd/main.go

# 2. Linux 64-bit Server / Mini PC (x86_64)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o bin/datalogger-linux-amd64 ./cmd/main.go

# 3. Raspberry Pi 4 / 5 — 64-bit OS (ARM64)
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "-s -w" -o bin/datalogger-linux-arm64 ./cmd/main.go

# 4. Raspberry Pi 2 / 3 — 32-bit OS (ARMv7)
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -ldflags "-s -w" -o bin/datalogger-linux-armv7 ./cmd/main.go

# 5. macOS Apple Silicon (M1-M4) & Intel
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o bin/datalogger-darwin-arm64 ./cmd/main.go
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o bin/datalogger-darwin-amd64 ./cmd/main.go
```

*(Skrip `install-linux.sh`, `install-windows.ps1`, dan `install-darwin.sh` secara otomatis akan mendeteksi dan menggunakan binary di folder `bin/` tersebut tanpa perlu menginstal Go di mesin target!)*

---

## 11. Uninstall & Troubleshooting

### 11.1. Prosedur Uninstall Aman (Data Telemetri Tetap Aman)

- **Linux / Raspberry Pi:**
  ```bash
  sudo bash ./scripts/install-linux.sh --uninstall
  ```
  *(Anda akan ditanya apakah ingin mempertahankan atau menghapus folder data `/var/lib/datalogger`).*
- **Windows:**
  ```powershell
  powershell -ExecutionPolicy Bypass -File .\scripts\install-windows.ps1 -Uninstall
  ```
- **macOS:**
  ```bash
  sudo bash ./scripts/install-darwin.sh --uninstall
  ```

### 11.2. Kendala Umum & Solusi Cepat

| Gejala / Error | Penyebab Umum | Solusi |
| :--- | :--- | :--- |
| `Database connection attempt 1/10 failed: Access denied` | Username/password di `config.json` tidak cocok dengan MariaDB lokal | Sesuaikan `db_user` & `db_password` di `config.json` (atau `/opt/datalogger/config.json` pada Linux), lalu restart service. |
| `bind: address already in use (:8080)` | Port `8080` sudah dipakai aplikasi lain atau proses lama masih berjalan | Ubah `"port": "8081"` di `config.json` atau hentikan proses yang memakai port `8080` (`sudo lsof -i :8080` di Linux / `Get-NetTCPConnection -LocalPort 8080` di PowerShell). |
| Halaman Web UI menampilkan `404` / kosong | Binary dijalankan dari luar folder kerja sehingga `./web/dist` tidak ditemukan | Jalankan binary dari dalam folder instalasi, atau set environment variable `DATALOGGER_WEB_DIR=/path/lengkap/ke/web/dist`. |
| Gagal baca sensor RS485 (`permission denied` di Linux) | User service belum punya izin baca `/dev/ttyUSB*` | Jalankan `sudo bash ./scripts/setup-usb-udev.sh` atau tambahkan user ke grup `dialout` (`sudo usermod -aG dialout $USER`). |

