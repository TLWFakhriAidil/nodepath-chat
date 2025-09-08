package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func main() {
	fmt.Println("🔍 Railway Build Diagnostic Check")
	fmt.Println("================================")
	fmt.Println()

	// Check Go version
	fmt.Printf("Go Version: %s\n", runtime.Version())
	fmt.Printf("GOOS: %s\n", runtime.GOOS)
	fmt.Printf("GOARCH: %s\n", runtime.GOARCH)
	fmt.Println()

	// Check if we can build main components
	fmt.Println("🚀 Testing Core Builds...")

	// Test main server build
	fmt.Print("Building main server... ")
	cmd := exec.Command("go", "build", "-o", "test-server", "./cmd/server")
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS=linux")
	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ FAILED: %v\n", err)
	} else {
		fmt.Println("✅ SUCCESS")
		os.Remove("test-server")
	}

	// Test migration build
	fmt.Print("Building migration utility... ")
	cmd = exec.Command("go", "build", "-o", "test-migrate", "./debug/fix_production_schema.go")
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS=linux")
	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ FAILED: %v\n", err)
	} else {
		fmt.Println("✅ SUCCESS")
		os.Remove("test-migrate")
	}

	// Test railway migration runner build
	fmt.Print("Building railway migration runner... ")
	cmd = exec.Command("go", "build", "-o", "test-railway", "./debug/railway_migration_runner.go")
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS=linux")
	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ FAILED: %v\n", err)
	} else {
		fmt.Println("✅ SUCCESS")
		os.Remove("test-railway")
	}

	fmt.Println()

	// Check required files
	fmt.Println("📁 Checking Required Files...")
	requiredFiles := []string{
		"production_fix_jam_column.sql",
		"start-with-migration.sh",
		"templates",
		"static",
		"go.mod",
		"go.sum",
		"package.json",
		"Dockerfile",
	}

	for _, file := range requiredFiles {
		if _, err := os.Stat(file); err == nil {
			fmt.Printf("✅ Found: %s\n", file)
		} else {
			fmt.Printf("⚠️ Missing: %s\n", file)
		}
	}

	fmt.Println()

	// Check go.mod for potential issues
	fmt.Println("📦 Go Module Analysis...")
	cmd = exec.Command("go", "mod", "verify")
	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ go mod verify failed: %v\n", err)
	} else {
		fmt.Println("✅ go mod verify passed")
	}

	// Check for large files
	fmt.Println()
	fmt.Println("🔍 Checking for Large Files (>5MB)...")
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && info.Size() > 5*1024*1024 {
			fmt.Printf("⚠️ Large file: %s (%.2f MB)\n", path, float64(info.Size())/(1024*1024))
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
	}

	fmt.Println()
	fmt.Println("🎉 Diagnostic Complete!")
	fmt.Println()
	fmt.Println("💡 Common Railway Build Issues:")
	fmt.Println("   1. Memory limits exceeded during build")
	fmt.Println("   2. Build timeout (usually 10-15 minutes)")
	fmt.Println("   3. Missing environment variables")
	fmt.Println("   4. Docker context size too large")
	fmt.Println("   5. Network connectivity issues")
	fmt.Println()
	fmt.Println("🔧 Recommended Solutions:")
	fmt.Println("   1. Add .dockerignore to reduce context size")
	fmt.Println("   2. Use multi-stage builds (already implemented)")
	fmt.Println("   3. Set proper resource limits in railway.toml")
	fmt.Println("   4. Check Railway logs for specific error messages")
}