# 🔍 Analisis Codebase — Genitz (go-initializr)

> Analisis menyeluruh terhadap kode, tampilan TUI, dan standar industri.

---

## 🎨 A. Kelemahan Tampilan (TUI / Visual)

### 1. Typo di nama folder template — `templetes` bukan `templates`
- Path: `internal/generator/templetes/`
- Konsisten salah di semua: konstanta, `depsPath`, `templateRoot`
- **Dampak**: Professional credibility turun saat orang baca source code

### 2. Divider `colorDivider` terlalu gelap / tidak terlihat
- Warna `#2D1B69` (indigo sangat gelap) pada `StepSep` dan `Divider` hampir tidak kelihatan di terminal dark theme biasa
- Garis pemisah antar step jadi "invisible" — user bingung step nav-nya nyambung atau nggak

### 3. Step nav tidak ada visual "back navigation"
- User bisa ketik `b` di Review untuk balik, tapi **tidak ada petunjuk visual** di step lain bahwa bisa back
- Step yang sudah `Done` (✓) tidak bisa diklik/navigate mundur secara eksplisit di step selain Review

### 4. Architecture view: tidak ada deskripsi panjang / preview struktur folder
- Deskripsi sangat singkat, misalnya `"cmd/ · internal/ · pkg/  — idiomatic Go layout"`
- Tidak ada preview folder tree visual seperti yang sering ada di starter generator modern

### 5. Splash logo tidak center secara horizontal
- `splashLogoWidth = 64` di-hardcode, tidak mengikuti terminal width
- Di terminal yang lebih lebar, logo muncul rata kiri — terasa kurang polished

### 6. Compact header hardcode versi `v0.1.0`
- File: `splash.go` baris 119
- Versi di-hardcode string literal, seharusnya inject saat build via `ldflags`

### 7. Dependency view: tidak ada scrolling / pagination
- Jika daftar dependency bertambah banyak, seluruh list langsung tampil tanpa batas
- Tidak ada indikator "scroll" (e.g. `↑ 3 more above`) seperti di alat CLI profesional

### 8. Color palette `StepSep` (`#2D1B69`) tidak konsisten dengan palette lain
- Warna ini sangat berbeda dari warna divider lain yang dipakai (`colorDivider`), tidak harmonis
- `colorDivider` juga digunakan untuk divider di `RenderHeaderCompact` — double concern

### 9. Tidak ada animasi / spinner saat generate
- Proses scaffolding bisa lama (go mod init, go get, go mod tidy)
- Saat ini hanya print `fmt.Println` biasa ke stdout — tidak ada progress spinner
- Alat CLI profesional (cobra, bubbletea sendiri sudah punya `spinner` component) seharusnya dipakai

### 10. `viewDone()` terlalu singkat dan langsung hilang
- Hanya render 2 baris lalu langsung quit — user tidak sempat melihat
- Tidak ada delay atau "press any key" confirmation

---

## 🧱 B. Kelemahan Kualitas Kode (Go Best Practices)

### 11. Semua dependency di `go.mod` adalah `// indirect`
- Tidak ada satu pun dependency yang di-mark sebagai direct (`require` tanpa `// indirect`)
- Ini terjadi karena semua `require` di-manage otomatis tanpa explicit `go get` di project root
- **Dampak**: Developer lain kebingungan mana yang intentional dependency vs transitive

### 12. `templateRoot` menggunakan relative path — fragile
- `const templateRoot = "internal/generator/templetes"`
- Ini **akan gagal** jika binary dijalankan dari direktori yang berbeda (bukan project root)
- Standar industri: embed template files dengan `//go:embed` agar portable

### 13. Template files tidak di-embed — binary tidak portabel
- Saat ini menggunakan `os.ReadFile` dengan relative path
- Jika binary di-`go install` dan dijalankan dari path lain, semua template tidak ditemukan
- **Fix**: Gunakan `embed.FS` dengan `//go:embed internal/generator/templetes/**`

### 14. `strings.Title` sudah deprecated di Go 1.18+
- File: `generate.go` baris 364 — `strings.Title(dep.Name)`
- Harus diganti dengan `golang.org/x/text/cases` package

### 15. `go 1.25.0` di `go.mod` — versi Go yang tidak ada
- Go 1.25 belum dirilis (saat analisis). Ini akan menyebabkan warning/error di beberapa toolchain

### 16. Tidak ada unit test sama sekali
- Tidak ada file `*_test.go` di `internal/generator/` atau `internal/tui/`
- Hanya ada `fiber_test.go` di template feature/fiber
- **Standar industri**: minimal ada test untuk `validateFolder`, `validateModulePath`, `resolveArchitecture`, `GenerateNewProject`

### 17. `Model` struct di TUI bersifat public semua field-nya
- Semua field (`FolderInput`, `PkgInput`, `ArchCursor`, dll.) di-export (kapital)
- Padahal hanya digunakan di package `tui` dan `main.go` (via `finalModel.Done`)
- Hanya `Done` yang perlu public; sisanya seharusnya private

### 18. Error handling di `main.go` bisa lebih informatif
- `fmt.Printf("Error: %v\n", err)` terlalu generic
- Tidak ada konteks error apa yang terjadi (TUI crash? input invalid?)

### 19. `Architecture` struct di `tui` package tapi hanya digunakan oleh `generator`
- Struct `Architecture` di `architecture.go` harusnya ada di `generator` package (atau package terpisah `types`)
- Sekarang `generator` package import `tui` package — ini circular dependency risk dan tidak clean

### 20. `depsPath` di `dependencies.go` sebagai relative path format string
- `var depsPath = "internal/generator/templetes/feature/%s"` — sama masalahnya dengan `templateRoot`

---

## 📦 C. Kelemahan Starter Project (Generator Output)

### 21. Hanya 2 dari 5 arsitektur yang punya template
- `architectureCatalog` hanya punya `Microservice` dan `Clean Architecture`
- `Standard Layout`, `DDD`, `CLI Tool` ditampilkan di UI tapi akan error saat dipilih: `"architecture %q is not supported yet"`
- **User experience yang buruk**: opsi yang tidak bisa digunakan seharusnya tidak ditampilkan

### 22. Hanya 3 dari 6 dependency yang punya `TemplateDir`
- `Gin Gonic`, `GORM`, `Uber Zap` punya `TemplateDir: ""` — dependency dipasang tapi **tidak ada boilerplate kode**
- User memilih GORM tapi tidak dapat generated code untuk koneksi DB, migration, dsb.

### 23. Template `microservice` sangat minimal
- Hanya punya folder `cmd/`, `config/`, `internal/`, `migrations/`, `pkg/`, `proto/`
- Tidak ada file Go starter (`main.go`, `server.go`, dsb.) di dalam folder tersebut
- Perlu dicek apakah ada file `.go.templete` di dalamnya

### 24. Tidak ada `README.md` yang di-generate
- Proyek yang di-generate tidak mendapat `README.md`
- Standar industri: semua scaffold tool (create-react-app, Spring Initializr, etc.) generate README

### 25. Tidak ada `Makefile` yang di-generate
- Hasil scaffold tidak punya `Makefile` untuk `run`, `build`, `test`, `lint`
- Padahal project Genitz sendiri punya Makefile yang bagus

### 26. Tidak ada `.gitignore` yang di-generate
- Proyek hasil generate tidak punya `.gitignore`
- Minimal harus ada: `*.env`, `bin/`, `vendor/`, `.idea/`, `.DS_Store`

### 27. Tidak ada `Dockerfile` / `docker-compose` template
- Untuk template microservice khususnya, ini sangat penting
- Standar industri modern: semua microservice scaffold harus punya Docker support

### 28. Tidak ada `.github/workflows` CI template
- Tidak ada CI/CD pipeline template (GitHub Actions)
- Standar industri: scaffold tool seperti ini biasanya include basic CI workflow

### 29. Config file generated tidak ada validasi environment
- `config.go.templete` hanya generate struct kosong
- Tidak ada contoh validasi environment variable yang hilang saat startup

### 30. Tidak ada `air.toml` / hot-reload config yang di-generate
- Makefile project Genitz sendiri ada `dev: air` tapi tidak ada `.air.toml` dan konfigurasi ini tidak di-generate ke project output

---

## 🏗️ D. Kelemahan Standar Industri Umum

### 31. Tidak ada `CHANGELOG.md` dan `CONTRIBUTING.md`
- Project open-source yang baik harus punya panduan kontribusi

### 32. Tidak ada versioning strategy
- Binary tidak embed version info
- Seharusnya gunakan `ldflags` saat build: `go build -ldflags "-X main.version=v0.1.0"`

### 33. Tidak ada `goreleaser` atau release pipeline
- Untuk distribusi binary ke user, perlu `.goreleaser.yml`
- Supaya bisa `go install` atau download binary dari GitHub Releases

### 34. Makefile menggunakan `mkdir -p` yang tidak bekerja di Windows natively
- `mkdir -p $(BUILD_DIR)` — ini sintaks Unix/bash, tidak bekerja di Windows PowerShell
- Perlu cross-platform alternative atau gunakan `go build` dengan `-o` saja

### 35. `flow.txt` masih ada di repo — noise / tidak diperlukan
- File ini seperti catatan pribadi/brainstorming yang seharusnya tidak di-commit
- Seharusnya masuk ke `.gitignore` atau dihapus

### 36. `artefact.sh` tidak ada dokumentasi/komentar
- Script shell tanpa penjelasan tujuannya apa

---

## ✅ Prioritas Perbaikan (Urut dari Kritis)

| # | Item | Dampak |
|---|------|--------|
| 🔴 | Embed templates pakai `//go:embed` | Binary tidak jalan di luar dev env |
| 🔴 | Hilangkan/grey-out arsitektur yang belum tersedia | UX rusak saat user pilih |
| 🔴 | Lengkapi TemplateDir untuk Gin, GORM, Zap | Feature yang dijanjikan tapi kosong |
| 🔴 | Fix `strings.Title` → `cases.Title` | Deprecated, warning di build |
| 🟠 | Generate `README.md`, `.gitignore`, `Makefile` | Standar minimum scaffold |
| 🟠 | Tambah progress spinner saat generate | UX buruk tanpa feedback visual |
| 🟠 | Fix versi di `go.mod` (1.25.0 tidak ada) | Toolchain warning/error |
| 🟠 | Center logo sesuai terminal width | Polish visual |
| 🟡 | Tambah unit tests | Maintainability |
| 🟡 | Pindahkan `Architecture` struct ke `generator` | Clean architecture |
| 🟡 | Versioning via `ldflags` | Release management |
| 🟡 | Rename `templetes` → `templates` | Code cleanliness |
