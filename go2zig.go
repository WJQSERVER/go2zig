package go2zig

import (
	"fmt"
	"go/build/constraint"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"go2zig/internal/generator"
	"go2zig/internal/parser"
)

type GenerateConfig struct {
	API          string
	Output       string
	PackageName  string
	LibraryName  string
	RuntimeZig   string
	BridgeZig    string
	APIModule    string
	ImplModule   string
	DynamicBuild bool
}

func Generate(cfg GenerateConfig) error {
	if cfg.API == "" {
		return fmt.Errorf("api path is required")
	}
	if cfg.Output == "" {
		return fmt.Errorf("output path is required")
	}
	if err := validateGeneratedBuildTag(cfg.Output); err != nil {
		return err
	}
	api, err := parser.ParseFile(cfg.API)
	if err != nil {
		return err
	}
	if cfg.LibraryName == "" {
		cfg.LibraryName = generator.LibraryNameFromPath(cfg.Output)
	}
	genCfg := generator.Config{
		PackageName: cfg.PackageName,
		LibraryName: cfg.LibraryName,
		APIModule:   defaultString(cfg.APIModule, "api.zig"),
		ImplModule:  defaultString(cfg.ImplModule, "lib.zig"),
	}
	content, err := generator.Render(api, genCfg)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(cfg.Output), 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	if err := os.WriteFile(cfg.Output, content, 0o644); err != nil {
		return fmt.Errorf("write generated go file: %w", err)
	}
	if cfg.RuntimeZig != "" {
		if err := os.WriteFile(cfg.RuntimeZig, generator.RenderZigRuntime(api, genCfg), 0o644); err != nil {
			return fmt.Errorf("write runtime zig file: %w", err)
		}
	}
	if cfg.BridgeZig != "" {
		if err := os.WriteFile(cfg.BridgeZig, generator.RenderZigBridge(api, genCfg), 0o644); err != nil {
			return fmt.Errorf("write bridge zig file: %w", err)
		}
	}
	return nil
}

func validateGeneratedBuildTag(outputPath string) error {
	content, err := os.ReadFile(outputPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read existing generated go file: %w", err)
	}
	tag, err := parseFirstBuildTag(string(content))
	if err != nil || tag == "" {
		return nil
	}
	expr, err := constraint.Parse(tag)
	if err != nil {
		expr, err = constraint.Parse("//go:build " + tag)
	}
	if err != nil {
		return nil
	}
	if !expr.Eval(func(tag string) bool {
		return tag == runtime.GOOS || tag == runtime.GOARCH
	}) {
		return fmt.Errorf("existing generated file %s is excluded on %s/%s; regenerate it with go2zig on a matching target so it gets the correct build tags, or remove the stale file", outputPath, runtime.GOOS, runtime.GOARCH)
	}
	return nil
}

func parseFirstBuildTag(content string) (string, error) {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "//go:build ") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "//go:build ")), nil
		}
		if strings.HasPrefix(trimmed, "//") {
			continue
		}
		break
	}
	return "", nil
}

type Builder struct {
	apiPath       string
	zigPath       string
	outputPath    string
	packageName   string
	libraryName   string
	optimize      string
	headerPath    string
	runtimeZig    string
	bridgeZig     string
	dynamicBuild  bool
	apiModuleName string
	implModule    string
}

func NewBuilder() *Builder {
	return &Builder{optimize: "ReleaseSafe", dynamicBuild: true}
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

func (b *Builder) WithRuntimeZig(path string) *Builder {
	b.runtimeZig = path
	return b
}

func (b *Builder) WithBridgeZig(path string) *Builder {
	b.bridgeZig = path
	return b
}

func (b *Builder) WithDynamicBuild(enabled bool) *Builder {
	b.dynamicBuild = enabled
	return b
}

func (b *Builder) WithAPIModuleName(name string) *Builder {
	b.apiModuleName = name
	return b
}

func (b *Builder) WithImplModule(name string) *Builder {
	b.implModule = name
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
	outputDir := filepath.Dir(b.outputPath)
	runtimeZig := b.runtimeZig
	if runtimeZig == "" {
		runtimeZig = filepath.Join(outputDir, "go2zig_runtime.zig")
	}
	bridgeZig := b.bridgeZig
	if bridgeZig == "" {
		bridgeZig = filepath.Join(outputDir, "go2zig_exports.zig")
	}
	apiModule := b.apiModuleName
	if apiModule == "" {
		apiModule = filepath.Base(apiPath)
	}
	implModule := b.implModule
	if implModule == "" && b.zigPath != "" {
		implModule = filepath.Base(b.zigPath)
	}
	if implModule == "" {
		implModule = "lib.zig"
	}

	if err := Generate(GenerateConfig{
		API:          apiPath,
		Output:       b.outputPath,
		PackageName:  b.packageName,
		LibraryName:  libraryName,
		RuntimeZig:   runtimeZig,
		BridgeZig:    bridgeZig,
		APIModule:    apiModule,
		ImplModule:   implModule,
		DynamicBuild: b.dynamicBuild,
	}); err != nil {
		return err
	}
	if b.zigPath == "" {
		return nil
	}
	return buildZig(b.zigPath, bridgeZig, b.outputPath, libraryName, b.optimize, b.headerPath, b.dynamicBuild)
}

func buildZig(zigPath, bridgePath, outputPath, libraryName, optimize, headerPath string, dynamic bool) error {
	if optimize == "" {
		optimize = "ReleaseSafe"
	}
	zigAbs, err := filepath.Abs(zigPath)
	if err != nil {
		return fmt.Errorf("resolve zig path: %w", err)
	}
	bridgeAbs, err := filepath.Abs(bridgePath)
	if err != nil {
		return fmt.Errorf("resolve bridge path: %w", err)
	}
	outputAbs, err := filepath.Abs(outputPath)
	if err != nil {
		return fmt.Errorf("resolve output path: %w", err)
	}
	outputDir := filepath.Dir(outputAbs)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create generated dir: %w", err)
	}
	binPath := filepath.Join(outputDir, outputLibraryFilename(libraryName, dynamic))
	args := []string{"build-lib"}
	if dynamic {
		args = append(args, "-dynamic")
	}
	args = append(args, "-O", optimize, "-femit-bin="+binPath)
	if headerPath != "" {
		headerPath, err = filepath.Abs(headerPath)
		if err != nil {
			return fmt.Errorf("resolve header path: %w", err)
		}
		args = append(args, "-femit-h="+headerPath)
	}
	args = append(args, filepath.Base(bridgeAbs))
	cmd := exec.Command("zig", args...)
	cmd.Dir = filepath.Dir(zigAbs)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("zig build-lib failed: %w", err)
	}
	return nil
}

func outputLibraryFilename(name string, dynamic bool) string {
	clean := strings.TrimSpace(name)
	if clean == "" {
		clean = "go2zig"
	}
	if dynamic {
		return generator.DynamicLibraryFilename(clean, runtime.GOOS)
	}
	if runtime.GOOS == "windows" {
		return clean + ".lib"
	}
	return "lib" + clean + ".a"
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
