package astparser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestASTParser(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "main.go")

	code := `package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	err := AddImport(filePath, "github.com/gin-gonic/gin")
	if err != nil {
		t.Fatalf("AddImport failed: %v", err)
	}

	err = InjectToMain(filePath, "r := gin.Default()\nr.Run()")
	if err != nil {
		t.Fatalf("InjectToMain failed: %v", err)
	}

	content, _ := os.ReadFile(filePath)
	str := string(content)

	if !strings.Contains(str, `"github.com/gin-gonic/gin"`) {
		t.Errorf("Import missed. File contains:\n%s", str)
	}
	if !strings.Contains(str, "r := gin.Default()") {
		t.Errorf("Injection missed. File contains:\n%s", str)
	}
}

func TestInjectStructField(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "config.go")

	code := `package config
type AppConfig struct {
	AppName string
	Port    int
}
`
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	err := InjectStructField(filePath, "AppConfig", "RedisHost string `json:\"redis_host\"`")
	if err != nil {
		t.Fatalf("InjectStructField failed: %v", err)
	}

	content, _ := os.ReadFile(filePath)
	str := string(content)

	if !strings.Contains(str, "RedisHost string `json:\"redis_host\"`") {
		t.Errorf("Struct Injection missed. File contains:\n%s", str)
	}
}
