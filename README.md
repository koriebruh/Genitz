# ⚡ Genitz CLI (Go Initializr)
> A next-generation, industrial-grade Go project scaffolding and package manager. 

**Genitz** beranjak jauh melampaui generator konvensional (seperti *Spring Initializr*). Dengan ditenagai oleh **AST (Abstract Syntax Tree) Engine**, Genitz tidak sekadar mencetak *file text template* kosong, tetapi mampu meretas *source code* secara organik, menyuntikkan konfigurasi struct, mengunduh file, dan mensimulasikan lingkungan auto-test per setiap instalasi.

---

## 🌟 Fitur Unggulan

- **Interactive Wizard TUI**: Antarmuka berkelas terminal dibalur komponen *BubbleTea* yang cantik, mulus, responsif, dan adaptif terhadap ukuran layar. Lengkap dengan _search bar_ dan tab arsitektur!
- **Zero-Break AST Injector**: Menyembunyikan kompleksitas modifikasi *go parsing tree*. Inisialisasi library (DB, framework, log, dll) dapat direkatkan secara _seamless_ ke kode buatan Anda tanpa `syntax error`.
- **Smart Config Struct Merger**: Menambahkan *property JSON tags* ke *struct* golang (misal `type Config struct`) secara otomatis kapan pun dependensi membutuhkannya.
- **BYOT (Bring Your Own Template)**: Jangan terkurung dengan struktur buatan Genitz. Pasang dan bangun template privat tim Git Anda sendiri hanya melalui *command line arguments*.
- **Auto-Mock Test Scaffold**: Jangan khawatir soal _coverage_ tes, karena Genitz merefleksikan file *table-driven test* secara otentik setiap kamu menarik *package*! 
- **`.env` Smart Merger**: Saat `genitz add` menambahkan dependensi baru (mis. Redis), variabel env baru ditambahkan ke file `.env` — tanpa pernah menimpa nilai yang sudah ada.
- **`genitz remove`**: Cabut library yang salah diinstall, import & kode init-nya ikut dihapus otomatis dari `main.go`.
- **Katalog 35+ Library Enterprise**: JWT, Casbin, Viper, Prometheus, OpenTelemetry, Asynq, Kafka, RabbitMQ, goose, Logrus, Zerolog, Echo, Chi, Sentry, dan masih banyak lagi. 

---

## 🚀 Instalasi Tercepat

1. **Clone** repositori ini ke lokal environtment Anda.
2. Compile _entry script_ Go-nya menjadi sistem Binary:
```bash
go build -o genitz main.go

# (Opsional) Pindahkan `genitz` (.exe) ke direktori $PATH system lokalmu agar bisa diakses di mana saja.
```

---

## 📚 Panduan Penggunaan Singkat

Mulai hari ini, Genitz tidak hanya dipakai di awal proyekmu, tapi akan memandumu mendirikan *business logic* mutakhir hingga tahap rilis!

### 1. Mode Terminal Wizard GUI (Membuat Project Baru)
Jalankan command utamanya tanpa parameter apapun untuk meluncurkan mode UI TUI BubbleTea:
```bash
genitz
```
Jendela modern yang atraktif akan muncul dan meminta input:
- Nama & Folder Project
- Arsitektur Base Code (Standard, Microservice, Hexagonal)
- Katalog Ratusan Dependency (Gin, Gorm, Fiber, SQLx, Zap, Viper, Testify, dst) dengan dukungan *Auto-Search*. 

Setelah menekan Generate, Genitz men-download dan memasangkan *snippet* kode secara pintar sembari menampilkan *live loading spinner*!

---

### 2. Mode Headless (SUNTIKAN PACKAGE TENGAH JALAN)
Jika di tengah pengembangan Anda menyadari lupa memasang `redis` atau `validator`, tak perlu pusing! Genitz punya mode **CLI Manager**:

```bash
# Tambah dependency baru
genitz add redis

# Cabut dependency yang salah diinstall
genitz remove redis
# atau shorthand:
genitz rm redis
```

Genitz akan secara gaib (*Silent Headless Mode*):
- Memeriksa keabsahan `go.mod` Anda.
- Mendownload repo pihak ketiga dengan `go get ...`.
- Mengeksekusi **Engine AST** untuk menyusup diam-diam, mencari blok `func main()`, dan menyuntikkan logika init (atau mencabutnya untuk `remove`).
- Menambahkan variabel ke `.env` secara non-destructive (tidak pernah menimpa nilai yang ada!).
- Menyusun _scaffolding template_ Mock Unit Test-nya!

**Dependency yang tersedia:**
| Kategori | Package ID |
|---|---|
| Framework | `fiber`, `gin`, `echo`, `chi` |
| ORM | `gorm`, `gorm-postgres`, `gorm-mysql`, `gorm-sqlite`, `gorm-sqlserver` |
| Driver (Native) | `pgx`, `mysql`, `mssqldb`, `clickhouse` |
| Cache | `redis` |
| Auth | `jwt`, `casbin` |
| Config | `viper` |
| Migration | `goose`, `migrate` |
| Logging | `zap`, `logrus`, `zerolog` |
| Observability | `prometheus`, `otel`, `sentry` |
| Background Jobs | `asynq`, `cron` |
| Messaging | `kafka`, `rabbitmq` |
| Utilities | `validator` |

---

### 3. Mode Peminjaman Template (BYOT)
Jika bosan dengan arsitektur default dan ingin merakit Base Golang buatan teman kantormu yang disimpan di Github:
```bash
genitz clone https://github.com/my-perusahann/go-scaffold-1.2 my-app
```
Detik berikutnya, repositori raksasa itu akan mendarat di komputer Anda (dengan _history_ `.git` yang sudah dipenggal rapi agar diisolasi jadi proyek lokal yang segar), beserta perombakan file `go.mod` menjadi `module my-app` instan. 

---

_Powered by BubbleTea, Charm, & Native Go Compiler Parser (AST)._
