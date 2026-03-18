package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- Styles ---
var (
	titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ADD8")).Bold(true).MarginBottom(1)

	// Style buat kursor (tanda panah ">")
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ADD8")).Bold(true)

	// Style buat Nama Dependency (selalu Bold & Putih/Terang)
	nameStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))

	// Style khusus pas item lagi dipilih (highlight)
	selectedNameStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00ADD8"))

	descStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Italic(true)
	checkStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")).Bold(true)
)

// --- Categories ---
const (
	CatFramework     = "framework"
	CatORM           = "orm"
	CatDriver        = "driver"
	CatCache         = "cache"
	CatMessageBroker = "broker"
	CatRPC           = "rpc"
	CatLogger        = "logger"
	CatTracing       = "tracing"
	CatMetrics       = "metrics"
	CatAuth          = "auth"
	CatValidation    = "validation"
	CatDoc           = "documentation"
)

func getBadgeStyle(category string) lipgloss.Style {
	base := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1).
		MarginLeft(1).
		Foreground(lipgloss.Color("#FFFFFF")) // Teks putih biar kontras

	switch category {
	case CatFramework:
		return base.Background(lipgloss.Color("#00ADD8")) // Gopher Blue
	case CatORM:
		return base.Background(lipgloss.Color("#F7931E")) // Orange (Database-ish)
	case CatDriver:
		return base.Background(lipgloss.Color("#4DB33D")) // Leaf Green (MongoDB/SQL)
	case CatCache:
		return base.Background(lipgloss.Color("#D82C20")) // Redis Red
	case CatMessageBroker:
		return base.Background(lipgloss.Color("#004E7A")) // Deep Blue (Kafka/NATS)
	case CatRPC:
		return base.Background(lipgloss.Color("#00B5AD")) // Teal (gRPC)
	case CatLogger:
		return base.Background(lipgloss.Color("#555555")) // Grey (Logs)
	case CatTracing:
		return base.Background(lipgloss.Color("#6B4E90")) // Purple (Observability)
	case CatMetrics:
		return base.Background(lipgloss.Color("#FF4500")) // Orange Red (Prometheus)
	case CatAuth:
		return base.Background(lipgloss.Color("#E91E63")) // Pink/Magenta (Security)
	case CatValidation:
		return base.Background(lipgloss.Color("#8BC34A")) // Light Green
	case CatDoc:
		return base.Background(lipgloss.Color("#3F51B5")) // Indigo (Swagger)
	default:
		return base.Background(lipgloss.Color("#222222")) // Dark Grey
	}
}

// --- Types ---
type Dependency struct {
	ID          string
	Name        string
	Category    string
	ImportPath  string
	IsDefault   bool
	Requires    []string
	Description string
}

type Model struct {
	Registry []Dependency     // Daftar master
	Chosen   map[int]struct{} // Index yang dicentang user
	Cursor   int              // Posisi kursor
	Done     bool             // Status apakah sudah enter
}

// --- Data ---
var DependencyRegistry = []Dependency{
	{
		ID: "gin", Name: "Gin Gonic", Category: CatFramework,
		ImportPath:  "github.com/gin-gonic/gin",
		Description: "High-performance HTTP web framework",
	},
	{
		ID: "gorm", Name: "GORM", Category: CatORM,
		ImportPath:  "gorm.io/gorm",
		Description: "The fantastic ORM library for Golang",
	},
	{
		ID: "zap", Name: "Uber Zap", Category: CatLogger,
		ImportPath:  "go.uber.org/zap",
		Description: "Blazing fast, structured, leveled logging",
	},
}

// --- Tea Functions ---
func InitialModel() Model {
	return Model{
		Registry: DependencyRegistry,
		Chosen:   make(map[int]struct{}),
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}

		case "down", "j":
			if m.Cursor < len(m.Registry)-1 {
				m.Cursor++
			}

		case " ": // Gunakan spasi untuk memilih (multi-select)
			_, ok := m.Chosen[m.Cursor]
			if ok {
				delete(m.Chosen, m.Cursor)
			} else {
				m.Chosen[m.Cursor] = struct{}{}
			}

		case "enter":
			m.Done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.Done {
		return "\n✅ Pilihan disimpan! Memproses generator...\n"
	}

	var s strings.Builder

	s.WriteString(titleStyle.Render("🚀 GO-INITIALIZR: Pilih Dependencies"))
	s.WriteString("\n")

	for i, dep := range m.Registry {
		// 1. Render Kursor
		cursor := "  "
		if m.Cursor == i {
			cursor = cursorStyle.Render("> ")
		}

		// 2. Render Checkbox
		checked := " [ ] "
		if _, ok := m.Chosen[i]; ok {
			checked = checkStyle.Render(" [x] ")
		}

		// 3. Render Nama Library (BOLD & KONTRAS)
		var name string
		if m.Cursor == i {
			// Pas kursor di sini, kita kasih warna biru cerah biar beda
			name = selectedNameStyle.Render(dep.Name)
		} else {
			// Pas kursor nggak di sini, tetep Bold tapi putih
			name = nameStyle.Render(dep.Name)
		}

		// 4. Render Badge Kategori
		badge := getBadgeStyle(dep.Category).Render(strings.ToUpper(dep.Category))

		// Gabungin baris utama
		s.WriteString(fmt.Sprintf("%s%s%s%s\n", cursor, checked, name, badge))

		// 5. Deskripsi
		s.WriteString(fmt.Sprintf("      %s\n", descStyle.Render(dep.Description)))
	}

	s.WriteString("\n(Space: Pilih, Enter: Lanjut, Q: Keluar)\n")
	return s.String()
}
