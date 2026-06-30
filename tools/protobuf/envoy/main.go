// Command envoy generates the Envoy grpc-web proxy config (envoy/launch.yaml)
// from the generated per-service OpenAPI specs under protobuf/generated/openapi.
//
// freebusy serves every gRPC service from a single backend, so the proxy needs
// one route per gRPC service — each matching the service's wire prefix
// "/<package>.<Service>/" — all pointing at one cluster. The merged
// openapiv3-spec.yaml only carries REST paths, not gRPC service names, so this
// tool reads the per-service *_service.openapi.yaml files instead: the package
// comes from each file's directory tree (freebusy/<domain>/v1 -> freebusy.<domain>.v1)
// and the service name from its operationId prefix ("<Service>_<Method>").
//
// Usage:
//
//	go run ./tools/protobuf/envoy [spec-dir] [flags]
//
// spec-dir defaults to protobuf/generated/openapi under the repo root; the
// output defaults to envoy/launch.yaml.
package main

import (
	_ "embed"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"gopkg.in/yaml.v3"
)

//go:embed launch.yaml.tmpl
var launchTemplate string

var verbose bool

// service is one gRPC service to route, e.g. freebusy.promocode.v1.PromoCodeService.
type service struct {
	Package string // freebusy.promocode.v1
	Name    string // PromoCodeService
	Prefix  string // /freebusy.promocode.v1.PromoCodeService/
}

// model is the data the launch.yaml template renders against.
type model struct {
	Source      string
	ListenPort  int
	AdminPort   int
	ClusterName string
	BackendHost string
	BackendPort int
	Timeout     string
	Services    []service
}

// serviceDoc captures just the operationIds from a per-service OpenAPI file;
// every other field is ignored.
type serviceDoc struct {
	Paths map[string]map[string]struct {
		OperationID string `yaml:"operationId"`
	} `yaml:"paths"`
}

func main() {
	root := repoRoot()
	defaultSpecDir := filepath.Join(root, "protobuf", "generated", "openapi")
	defaultOutput := filepath.Join(root, "docker/envoy", "launch.yaml")

	var (
		output      string
		listenPort  int
		adminPort   int
		backendHost string
		backendPort int
		cluster     string
		timeout     string
	)
	flag.StringVar(&output, "output", defaultOutput, "Path to write the Envoy config")
	flag.StringVar(&output, "o", defaultOutput, "Path to write the Envoy config (shorthand)")
	flag.IntVar(&listenPort, "listen-port", 8080, "Port the grpc-web listener binds")
	flag.IntVar(&adminPort, "admin-port", 9901, "Port the Envoy admin UI binds")
	flag.StringVar(&backendHost, "backend-host", "host.docker.internal",
		"Hostname of the freebusy gRPC backend (use 127.0.0.1 when running Envoy natively)")
	flag.IntVar(&backendPort, "backend-port", 50051, "Port of the freebusy gRPC backend")
	flag.StringVar(&cluster, "cluster", "freebusy_backend", "Name of the backend cluster")
	flag.StringVar(&timeout, "timeout", "60s", "Per-route upstream timeout")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose (debug) logging")
	flag.BoolVar(&verbose, "v", false, "Enable verbose (debug) logging (shorthand)")
	flag.Parse()

	specDir := defaultSpecDir
	if flag.NArg() > 0 {
		specDir = flag.Arg(0)
	}
	specDir, _ = filepath.Abs(specDir)
	if fi, err := os.Stat(specDir); err != nil || !fi.IsDir() {
		fatalf("spec directory does not exist: %s", specDir)
	}

	services := collectServices(specDir)
	if len(services) == 0 {
		fatalf("no gRPC services found under %s (expected *_service.openapi.yaml files)", specDir)
	}

	source := "protobuf/generated/openapi/**/*_service.openapi.yaml"
	if rel, err := filepath.Rel(root, specDir); err == nil {
		source = filepath.ToSlash(filepath.Join(rel, "**", "*_service.openapi.yaml"))
	}

	data := model{
		Source:      source,
		ListenPort:  listenPort,
		AdminPort:   adminPort,
		ClusterName: cluster,
		BackendHost: backendHost,
		BackendPort: backendPort,
		Timeout:     timeout,
		Services:    services,
	}

	output, _ = filepath.Abs(output)
	if err := render(output, data); err != nil {
		fatalf("write output: %v", err)
	}
	logf("INFO", "envoy config written to %s (%d services -> cluster %q)", output, len(services), cluster)
}

// collectServices reads every *_service.openapi.yaml under specDir and returns
// the gRPC services to route, de-duplicated by full name and sorted by prefix.
func collectServices(specDir string) []service {
	files := findFiles(specDir, "_service.openapi.yaml")
	if len(files) == 0 {
		logf("WARN", "no *_service.openapi.yaml files found under %s", specDir)
	}

	seen := map[string]bool{}
	var services []service
	for _, file := range files {
		rel, _ := filepath.Rel(specDir, file)
		pkg := packageFromPath(filepath.Dir(rel))
		if pkg == "" {
			logf("WARN", "skipping %s: cannot derive package from path", rel)
			continue
		}
		name := serviceName(file)
		if name == "" {
			logf("WARN", "skipping %s: no operationId to derive a service name from", rel)
			continue
		}
		full := pkg + "." + name
		if seen[full] {
			continue
		}
		seen[full] = true
		services = append(services, service{
			Package: pkg,
			Name:    name,
			Prefix:  "/" + full + "/",
		})
		logf("DEBUG", "routing %s (from %s)", full, rel)
	}

	sort.Slice(services, func(i, j int) bool { return services[i].Prefix < services[j].Prefix })
	return services
}

// packageFromPath turns a relative directory like "freebusy/promocode/v1" into
// the proto package "freebusy.promocode.v1".
func packageFromPath(dir string) string {
	dir = filepath.ToSlash(dir)
	if dir == "" || dir == "." {
		return ""
	}
	return strings.ReplaceAll(dir, "/", ".")
}

// serviceName reads the first operationId from a service spec and returns the
// "<Service>" portion of its "<Service>_<Method>" convention.
func serviceName(file string) string {
	data, err := os.ReadFile(file)
	if err != nil {
		logf("ERROR", "read %s: %v", file, err)
		return ""
	}
	var doc serviceDoc
	if err := yaml.Unmarshal(data, &doc); err != nil {
		logf("ERROR", "parse %s: %v", file, err)
		return ""
	}
	// Paths/methods iterate in map order; collect candidate names so the result
	// is deterministic regardless of which operation we see first.
	names := map[string]bool{}
	for _, item := range doc.Paths {
		for _, op := range item {
			if op.OperationID == "" {
				continue
			}
			if i := strings.IndexByte(op.OperationID, '_'); i > 0 {
				names[op.OperationID[:i]] = true
			}
		}
	}
	switch len(names) {
	case 0:
		return ""
	case 1:
		for n := range names {
			return n
		}
	}
	// A single file is expected to define one service; if it somehow spans
	// several, pick the lexically first so output stays stable, and warn.
	sorted := make([]string, 0, len(names))
	for n := range names {
		sorted = append(sorted, n)
	}
	sort.Strings(sorted)
	logf("WARN", "%s defines multiple services %v; using %q", filepath.Base(file), sorted, sorted[0])
	return sorted[0]
}

// render writes the launch.yaml template filled with data.
func render(path string, data model) (err error) {
	tmpl, err := template.New("launch.yaml").Parse(launchTemplate)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	return tmpl.Execute(f, data)
}

// --- helpers (mirrors tools/protobuf/openapi) ---

// findFiles returns every file under dir whose name ends with suffix, sorted.
func findFiles(dir, suffix string) []string {
	var files []string
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), suffix) {
			files = append(files, path)
		}
		return nil
	})
	sort.Strings(files)
	return files
}

// repoRoot walks up from the working directory to the directory holding go.mod
// so paths resolve regardless of where the command is invoked from.
func repoRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "."
		}
		dir = parent
	}
}

func logf(level, format string, args ...any) {
	if level == "DEBUG" && !verbose {
		return
	}
	fmt.Fprintf(os.Stderr, "%s %-7s %s\n", time.Now().Format("15:04:05"), level, fmt.Sprintf(format, args...))
}

func fatalf(format string, args ...any) {
	logf("ERROR", format, args...)
	os.Exit(1)
}
