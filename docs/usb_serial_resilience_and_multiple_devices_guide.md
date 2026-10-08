# Panduan Ketahanan USB Serial (RS485) & Penanganan Banyak Perangkat di Lapangan
> **Datalogger Edge System — Industrial Field Reliability Standard**  
> *Solusi Penanganan Port Berpindah (`ttyUSB0` ➔ `ttyUSB1`) & Pengelolaan Banyak Sensor Serial*

---

## 1. Latar Belakang & Masalah Nyata di Lapangan

Dalam instalasi industri dan monitoring lapangan berbasis Raspberry Pi atau Linux SBC, konverter **USB-ke-RS485/RS232** (seperti chip FTDI FT232R, Silicon Labs CP2102, Prolific PL2303, atau WCH CH340) merupakan komponen vital yang menghubungkan Datalogger dengan sensor Modbus RTU (KWH Meter, Weather Station, Sensor Suhu & Kelembaban, Inverter, Flow Meter).

Namun, ada dua masalah umum yang sering dihadapi teknisi lapangan:

### Masalah 1: Lompatan Nomor Port Serial (`ttyUSB0` ➔ `ttyUSB1`)
* **Penyebab:**
  1. **Induksi & ESD:** Spark atau medan magnet sesaat dari motor/inverter di panel boks.
  2. **Drop Tegangan (Brownout):** Fluktuasi daya 5V pada port USB saat perangkat lain menyala.
  3. **Getaran Mekanik:** Konektor USB yang agak kendor mengalami *micro-disconnect* (< 100ms).
* **Gejala di Linux:**
  Ketika USB terputus sesaat, kernel Linux mencatat disconnect. Namun proses aplikasi belum sempat menutup file descriptor `/dev/ttyUSB0`. Saat USB tersambung kembali beberapa milidetik kemudian, kernel menganggap nama `/dev/ttyUSB0` masih terpakai, sehingga mengalokasikan nomor urut berikutnya: **`/dev/ttyUSB1`** (lalu `ttyUSB2`, dst).
* **Akibatnya:** Datalogger yang dikonfigurasi membaca `/dev/ttyUSB0` mengalami error `serial port offline: The system cannot find the file specified`, dan pembacaan sensor terhenti.

---

### Masalah 2: Urutan Port Tertukar Ketika Banyak USB Serial Terhubung
* **Skenario:** Ada 3 konverter USB RS485 yang dicolok bersamaan ke Raspberry Pi:
  - USB A = Sensor KWH Meter Listrik
  - USB B = Sensor Suhu & Kelembaban Lingkungan
  - USB C = Inverter Surya
* **Penyebab:**
  Kernel Linux menginisialisasi perangkat USB saat boot secara asinkron (*race condition*). Chip USB yang merespons 5ms lebih cepat akan menjadi `/dev/ttyUSB0`, yang lambat menjadi `/dev/ttyUSB1`, dst.
* **Akibatnya:** Setelah Raspberry Pi mati lampu atau restart, **identitas port tertukar**:
  - KWH Meter terbaca di `/dev/ttyUSB1` (sebelumnya `ttyUSB0`).
  - Suhu terbaca di `/dev/ttyUSB0` (sebelumnya `ttyUSB1`).
  - Data yang disimpan ke database menjadi kacau atau komunikasi gagal total karena perbedaan baud rate / slave ID Modbus.

---

## 2. Solusi Komprehensif: Standar 3 Pilar Datalogger

Untuk mengatasi kedua masalah di atas secara permanen, Datalogger menerapkan arsitektur **3 Pilar Ketahanan Port**:

```
                  [ TANTANGAN DI LAPANGAN ]
          (Micro-Disconnect / Re-Enumerasi / Banyak USB)
                             │
       ┌─────────────────────┼─────────────────────┐
       ▼                     ▼                     ▼
 ┌───────────┐         ┌───────────┐         ┌───────────┐
 │  PILAR 1  │         │  PILAR 2  │         │  PILAR 3  │
 │Jalur Tetap│         │SelfHealing│         │udev Rules │
 │ Linux OS  │         │Software Go│         │ Otomatis  │
 └─────┬─────┘         └─────┬─────┘         └─────┬─────┘
       │                     │                     │
       ▼                     ▼                     ▼
• by-id/ (Chip Serial) • ResolvePortAddress  • setup-usb-udev.sh
• by-path/ (Colokan Pi)• Auto-Migrate tty0➔1 • Alias tetap di /dev/
• Zero Config          • Log & Rebind        • Panel Boks Pabrik
```

---

### PILAR 1: Menggunakan Jalur Permanen Linux Bawaan (`by-id` & `by-path`)

Linux sebenarnya telah menyediakan symlink permanen di folder `/dev/serial/`:

#### A. `/dev/serial/by-id/` (Berdasarkan Nomor Seri Chip USB)
* **Format:** `/dev/serial/by-id/usb-<Manufacturer>_<Model>_<ChipSerial>-if00-port0`
* **Contoh:** `/dev/serial/by-id/usb-FTDI_FT232R_USB_UART_A900abc1-if00-port0`
* **Cara Kerja:**
  Symlink ini mengikat langsung ke **nomor seri unik (iSerial)** yang tertanam di chip konverter USB. Mau kabel dicabut-colok 100 kali, atau melompat dari `ttyUSB0` ke `ttyUSB1`, symlink ini **OTOMATIS selalu diarahkan oleh kernel Linux ke perangkat aktif**.
* **Kapan Digunakan:** Sangat disarankan jika konverter USB menggunakan chip berkualitas seperti FTDI (FT232RL), Silicon Labs (CP2102), atau CH340N yang memiliki nomor seri unik.

#### B. `/dev/serial/by-path/` (Berdasarkan Colokan Fisik USB Raspberry Pi)
* **Format:** `/dev/serial/by-path/platform-<controller>-usb-0:<hub>.<port>:1.0-port0`
* **Contoh:** `/dev/serial/by-path/platform-fd500000.pcie-pci-0000:01:00.0-usb-0:1.3:1.0-port0`
* **Cara Kerja:**
  Symlink ini mengikat sensor berdasarkan **lubang colokan fisik USB di board Raspberry Pi**. 
  - Colokan Kiri Atas = Port 1.1
  - Colokan Kiri Bawah = Port 1.2
  - Colokan Kanan Atas = Port 1.3
  - Colokan Kanan Bawah = Port 1.4
* **Kapan Digunakan:** Sangat ideal jika Anda menggunakan **beberapa converter murah yang identik** (misal 3 adapter CH340 generik yang tidak memiliki nomor seri unik). Selama kabel tidak dipindah colokan fisiknya, jalur sensor 100% tidak akan pernah tertukar!

---

### PILAR 2: Fitur Self-Healing Rebinding di Backend Go

Jika teknisi lapangan terlanjur mendaftarkan port dengan nama dinamis `/dev/ttyUSB0`:

1. **Pendeteksian Otomatis:**
   Di modul [`internal/communication/modbus/rtu.go`](file:///d:/cbi-project-src/datalogger-27/internal/communication/modbus/rtu.go), fungsi `ResolvePortAddress` dan `handleSerialError` aktif mengawasi status koneksi.
2. **Kondisi 1 USB Terpasang:**
   Jika `/dev/ttyUSB0` tiba-tiba hilang dan kernel memunculkan `/dev/ttyUSB1`, backend Datalogger **otomatis mendeteksi lompatan re-enumerasi dan memindahkan pointer port ke `/dev/ttyUSB1` secara mandiri**.
3. **Pencatatan Audit Log:**
   Backend mencatat log resmi:
   ```log
   [WARN] USB serial port migration detected! Configured '/dev/ttyUSB0' is missing, but active '/dev/ttyUSB1' was found. Auto-rebinding transport to '/dev/ttyUSB1'
   ```
4. **Hasil:** Komunikasi sensor otomatis pulih dalam siklus berikutnya tanpa perlu me-restart service Datalogger dan tanpa perlu kehadiran teknisi di lokasi.

---

### PILAR 3: Generator Aturan udev Permanen (`scripts/setup-usb-udev.sh`)

Untuk instalasi panel boks permanen di pabrik atau substation, standar terbaik adalah membuat nama alias kustom yang jelas dan manusiawi, misalnya:
* `/dev/datalogger_rs485_1` ➔ KWH Meter Ruang Panel
* `/dev/datalogger_rs485_2` ➔ Sensor Suhu & Kelembaban Server

Datalogger telah menyediakan tool otomatis:
```bash
sudo bash scripts/setup-usb-udev.sh
```

#### Cara Kerja Script:
1. Script memindai seluruh adapter USB serial yang sedang tercolok.
2. Membaca Vendor ID, Product ID, Chip Serial, dan USB Tree Path secara otomatis via `udevadm`.
3. Menulis aturan udev ke `/etc/udev/rules.d/99-datalogger-serial.rules`.
4. Memicu reload aturan (`udevadm control --reload-rules && udevadm trigger`).
5. Menghasilkan symlink alias `/dev/datalogger_rs485_*` dengan izin akses non-root (`MODE="0666"`).

---

## 3. Panduan Praktis untuk Teknisi Lapangan

### Langkah Registrasi Sensor di Web Dashboard:
1. Buka menu **Devices & Sensors** (`#/monitoring/devices`).
2. Klik tombol **Register Device** atau klik ikon Edit pada perangkat yang ada.
3. Buka tab **Communication & Protocol**, pilih protocol `MODBUS_RTU` atau `SERIAL`.
4. Klik tombol **Scan Ports**:
   Dropdown akan secara otomatis mengelompokkan port:
   * 🔒 **Rekomendasi: Persistent by-ID** *(Pilih opsi ini jika ada)*
   * 📍 **Rekomendasi: Persistent by-Path** *(Pilih opsi ini jika menggunakan dongle identik)*
   * ⚡ **Direct Ports** *(Gunakan hanya untuk pengujian sementara)*
5. Simpan perangkat. Datalogger akan langsung berkomunikasi dengan sensor secara stabil.

---

## 4. Tabel Perbandingan Pilihan Port Serial

| Jenis Konfigurasi Port | Contoh Format | Tahan Glitch / Re-enumeration? | Tahan Reboot (Banyak USB)? | Rekomendasi Penggunaan |
|---|---|:---:|:---:|---|
| **Direct Port** | `/dev/ttyUSB0` | ⚠️ Butuh Self-Healing | ❌ Rentan Tertukar | Pengujian sementara di meja kerja |
| **Persistent by-ID** | `/dev/serial/by-id/usb-FTDI...` | ✅ Sangat Tahan | ✅ Sangat Aman | **Sangat Disarankan** di Lapangan |
| **Persistent by-Path** | `/dev/serial/by-path/platform...` | ✅ Sangat Tahan | ✅ Sangat Aman (kunci colokan) | **Sangat Disarankan** untuk dongle murah |
| **Custom udev Alias** | `/dev/datalogger_rs485_1` | ✅ Sangat Tahan | ✅ 100% Permanen | Panel Boks Produksi / Industri |

---

## 5. Troubleshooting Lapangan

### Pertanyaan 1: Bagaimana jika `/dev/serial/by-id/` tidak muncul sama sekali di Raspberry Pi?
Beberapa converter chip clone CH340 super murah tidak menyertakan deskriptor serial number USB.  
**Solusi:** Gunakan jalur **`/dev/serial/by-path/`**, yang pasti selalu ada di setiap Raspberry Pi karena jalur ini dibuat berdasarkan arsitektur sirkuit bus USB fisik motherboard.

### Pertanyaan 2: Bagaimana cara mengecek daftar port serial via terminal di Raspberry Pi?
Jalankan salah satu perintah berikut:
```bash
# Cek port direct:
ls -l /dev/ttyUSB* /dev/ttyACM*

# Cek symlink by-id:
ls -l /dev/serial/by-id/

# Cek symlink by-path:
ls -l /dev/serial/by-path/
```

### Pertanyaan 3: Muncul error `permission denied` saat membaca port serial di Raspberry Pi?
Pastikan user Linux Anda sudah masuk ke grup `dialout`:
```bash
sudo usermod -aG dialout $USER
# atau jalankan script udev datalogger yang otomatis memberikan MODE="0666"
sudo bash scripts/setup-usb-udev.sh
```
