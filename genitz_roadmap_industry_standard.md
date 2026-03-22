# 🚀 Roadmap Genitz: "Beyond Spring Initializr"

Untuk menjadikan `Genitz` (Go Initializr) setara atau bahkan mengalahkan **Spring Initializr** di industri (khususnya untuk ekosistem Go), aplikasi ini tidak cukup hanya dengan *copy-paste template* atau me-replace text. Aplikasi harus lebih cerdas, *modular*, dan mendukung skalabilitas jangka panjang.

Berikut adalah pilar utama perbaikan logika dan arsitektur yang perlu ditambahkan:

---

## 🛠️ Tahap 1: Smart Scaffolding (Lebih Pintar dari Template Teks)

Spring Initializr menang karena ia men-generate file `.pom` / `build.gradle` secara dinamis tanpa merusak deklarasi yang ada. Di Go, kita butuh ini pada file [go.mod](file:///d:/go-initializr/go.mod) dan [main.go](file:///d:/go-initializr/main.go).

### 1. **AST-Based Code Injection** (Sangat Kritis)
Saat ini generator Genitz menggabungkan template menggunakan string substitution (`{{.ConfigInit}}`). Ini rentan bentrok jika struktur berubah.
- **Implementasi Ideal:** Menggunakan package `go/ast` (Abstract Syntax Tree) dan `go/parser` milik Go. Alih-alih me-replace teks, Genitz harus "membaca" file [main.go](file:///d:/go-initializr/main.go), lalu *menyuntikkan* (inject) fungsi seperti [NewGin()](file:///d:/go-initializr/internal/generator/templates/feature/gin/gin.go#27-41), [NewDatabase()](file:///d:/go-initializr/internal/generator/templates/feature/gorm/gorm.go#30-55), atau blok impor baru secara algoritmik. Ini menjamin source code tidak pernah *syntax error* saat selesai di-generate.

### 2. **Dependency Resolution & Compatibility Matrix**
Spring Initializr tahu jika kamu pilih Java 11, *library X versi 3* tidak akan cocok.
- **Logic Genitz:** Perlu sistem *compatibility matrix*. Jika user memilih Go `1.23`, jangan meng-import library tipe lama. Jika user milih `GORM` + `Postgres`, Genitz harus otomatis tahu untuk menarik *driver postgres gorm*. 

### 3. **Smart Config Merging** 
Daripada men-generate `config.go` secara statis, buat logic untuk membaca [.env](file:///d:/go-initializr/internal/generator/templates/feature/gin/.env) lama dan menggabungkan struct `*Config` ke dalam satu struct global yang rapi otomatis.

---

## 🌐 Tahap 2: CLI & Web Duality (Standar Industri Modern)

### 4. **API Endpoint & Web UI**
Spring Initializr sukses karena punya UI web dan endpoint cURL.
- **Genitz Backend:** Core logic generator Genitz (package `generator`) harus murni dipisah dari TUI (`bubbletea`).
- **Genitz API Server:** Tambahkan kemampuan menjalankan `genitz serve` yang akan memunculkan server REST API. Contoh:
  ```bash
  curl -G https://start.genitz.io/starter.zip -d type=microservice -d deps=gin,gorm -o my-app.zip
  ```

### 5. **Remote Templates & Custom Schemes**
Perusahaan besar punya standar (clean architecture versi *lokal* mereka).
- Genitz harus bisa men-download spesifikasi template dari Git:
  `genitz --template github.com/my-company/go-template`
- Ini menjadikan Genitz sebagai **Universal Go Project Manager**, bukan cuma generator biasa.

---

## 🏗️ Tahap 3: Next-Gen "Project Manager" (Beyond Initializr)

Di titik ini, Genitz mengalahkan Spring Initializr yang hanya sebatas membuat *project dari awal*.

### 6. **Fitur "Add" pada Project Eksisting**
Jangan hanya [init](file:///d:/go-initializr/internal/generator/generate.go#331-350). Jika di tengah-tengah development orang butuh Kafka, mereka bisa masuk ke folder aplikasi lamanya dan menjalankan:
- `genitz add kafka`
- Maka Genitz (berbekal fitur AST nomor 1) akan otomatis memodifikasi [main.go](file:///d:/go-initializr/main.go) yang sudah ada, mensuntik inisialisasi kafka, dan menambahkan boilerplate logic-nya tanpa merusak kode *business logic* user.

### 7. **Instant Production-Ready (Ops in a Box)**
Begitu di-generate, project sudah di-inject dengan standar deployment mutakhir:
- **Distroless Multi-Stage Dockerfile** (Sangat tipis, aman standar kubernetes).
- **Helm Charts** otomatis sesuai nama project.
- **CI/CD GitHub Actions / GitLab CI** dengan caching `golangci-lint` otomatis.

### 8. **Built-in Mocking & Test Scaffolding**
Saat user memilih arsitektur, Genitz juga membuatkan:
- Struktur package interface untuk mempermudah `gomock`
- File [_test.go](file:///d:/go-initializr/internal/generator/templetes/feature/fiber/fiber_test.go) framework template (table-driven tests bawaan Go).

---

## 🧩 Ringkasan Arsitektur Genitz Masa Depan

Jika ingin Genitz di-scale, arsitektur kode internalnya harus diubah dari script prosedural menjadi:

1. **`core/parser`**: Sistem `AST` untuk membaca dan menyuntik syntax Go.
2. **`core/plugins`**: Setiap dependensi (contoh `zap`, `gorm`) adalah *plugin*. Masing-masing plugin punya method `InjectImports()`, `InjectMainInit()`, dan `WriteFiles()`.
3. **`pkg/tui`**: UI interaktif BubbleTea.
4. **`pkg/server`**: Web API server.
5. **`cmd/genitz`**: CLI entry point.

Jika Anda sanggup mengubah alur *templating* `text/template` lama ke sistem *Go AST Injection*, saat itulah Genitz benar-benar menjadi alat **Standard Industri Multi-Level** yang tidak tertandingi alat generator konvensional.
