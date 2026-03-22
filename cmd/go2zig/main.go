package main

import (
	"flag"
	"fmt"
	"os"

	"go2zig"
)

func main() {
	var cfg struct {
		api     string
		zig     string
		out     string
		pkg     string
		lib     string
		opt     string
		header  string
		runtime string
		bridge  string
		noBuild bool
	}
	flag.StringVar(&cfg.api, "api", "", "path to zig api declaration file")
	flag.StringVar(&cfg.zig, "zig", "", "path to zig library source to compile")
	flag.StringVar(&cfg.out, "out", "", "destination go file")
	flag.StringVar(&cfg.pkg, "pkg", "main", "generated go package name")
	flag.StringVar(&cfg.lib, "lib", "", "library name used by cgo")
	flag.StringVar(&cfg.opt, "opt", "ReleaseSafe", "zig optimization mode")
	flag.StringVar(&cfg.header, "header", "", "optional generated header path")
	flag.StringVar(&cfg.runtime, "runtime-zig", "", "generated zig runtime helper file")
	flag.StringVar(&cfg.bridge, "bridge-zig", "", "generated zig export bridge file")
	flag.BoolVar(&cfg.noBuild, "no-build", false, "only generate go wrapper without compiling zig")
	flag.Parse()

	b := go2zig.NewBuilder().
		WithAPI(cfg.api).
		WithOutput(cfg.out).
		WithPackageName(cfg.pkg).
		WithLibraryName(cfg.lib).
		WithOptimize(cfg.opt).
		WithHeaderOutput(cfg.header).
		WithRuntimeZig(cfg.runtime).
		WithBridgeZig(cfg.bridge)
	if !cfg.noBuild {
		b.WithZigSource(cfg.zig)
	}
	if err := b.Build(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
