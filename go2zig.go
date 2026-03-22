package go2zig

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"go2zig/internal/generator"
	"go2zig/internal/parser"
)

type GenerateConfig struct {
	API         string
	Output      string
	PackageName string
	LibraryName string
}

func Generate(cfg GenerateConfig) error {
	if cfg.API == "" {
		return fmt.Errorf("api path is required")
	}
	if cfg.Output == "" {
		return fmt.Errorf("output path is required")
	}
	api, err := parser.ParseFile(cfg.API)
	if err != nil {
		return err
	}
	if cfg.LibraryName == "" {
		cfg.LibraryName = generator.LibraryNameFromPath(cfg.Output)
	}
	content, err := generator.Render(api, generator.Config{
		PackageName: cfg.PackageName,
		LibraryName: cfg.LibraryName,
	})
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(cfg.Output), 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	if err := os.WriteFile(cfg.Output, content, 0o644); err != nil {
		return fmt.Errorf("write generated go file: %w", err)
	}
	return nil
}

type Builder struct {
	apiPath     string
	zigPath     string
	outputPath  string
	packageName string
	libraryName string
	optimize    string
	headerPath  string
}

func NewBuilder() *Builder {
	return &Builder{optimize: "ReleaseSafe"}
}

func (b *Builder) WithAPI(path string) *Builder {
	b.apiPath = path
	return b
}

func (b *Builder) WithZigSource(path string) *Builder {
	b.zigPath = path
	return b
}

func (b *Builder) WithOutput(path string) *Builder {
	b.outputPath = path
	return b
}

func (b *Builder) WithPackageName(name string) *Builder {
	b.packageName = name
	return b
}

func (b *Builder) WithLibraryName(name string) *Builder {
	b.libraryName = name
	return b
}

func (b *Builder) WithOptimize(mode string) *Builder {
	b.optimize = mode
	return b
}

func (b *Builder) WithHeaderOutput(path string) *Builder {
	b.headerPath = path
	return b
}

func (b *Builder) Build() error {
	if b.outputPath == "" {
		return fmt.Errorf("output path is required")
	}
	if b.apiPath == "" && b.zigPath == "" {
		return fmt.Errorf("api path or zig source path is required")
	}

	apiPath := b.apiPath
	if apiPath == "" {
		apiPath = b.zigPath
	}
	libraryName := b.libraryName
	if libraryName == "" {
		if b.zigPath != "" {
			libraryName = generator.LibraryNameFromPath(b.zigPath)
		} else {
			libraryName = generator.LibraryNameFromPath(b.outputPath)
		}
	}

	if err := Generate(GenerateConfig{
		API:         apiPath,
		Output:      b.outputPath,
		PackageName: b.packageName,
		LibraryName: libraryName,
	}); err != nil {
		return err
	}
	if b.zigPath == "" {
		return nil
	}
	return buildZig(b.zigPath, b.outputPath, libraryName, b.optimize, b.headerPath)
}

func buildZig(zigPath, outputPath, libraryName, optimize, headerPath string) error {
	if optimize == "" {
		optimize = "ReleaseSafe"
	}
	zigAbs, err := filepath.Abs(zigPath)
	if err != nil {
		return fmt.Errorf("resolve zig path: %w", err)
	}
	outputAbs, err := filepath.Abs(outputPath)
	if err != nil {
		return fmt.Errorf("resolve output path: %w", err)
	}
	outputDir := filepath.Dir(outputAbs)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create generated dir: %w", err)
	}
	libPath := filepath.Join(outputDir, staticLibraryFilename(libraryName))
	args := []string{"build-lib", "-lc", "-O", optimize, "-femit-bin=" + libPath}
	if headerPath != "" {
		headerPath, err = filepath.Abs(headerPath)
		if err != nil {
			return fmt.Errorf("resolve header path: %w", err)
		}
		args = append(args, "-femit-h="+headerPath)
	}
	args = append(args, filepath.Base(zigAbs))
	cmd := exec.Command("zig", args...)
	cmd.Dir = filepath.Dir(zigAbs)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("zig build-lib failed: %w", err)
	}
	return nil
}

func staticLibraryFilename(name string) string {
	clean := strings.TrimSpace(name)
	if clean == "" {
		clean = "go2zig"
	}
	if runtime.GOOS == "windows" {
		return "lib" + clean + ".a"
	}
	return "lib" + clean + ".a"
}
