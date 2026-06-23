// Command openapi merges the generated per-service OpenAPI specs under
// protobuf/generated/openapi into a single openapiv3-spec.yaml.
//
// It loads every *.openapi.yaml (paths + "weak" component definitions), overlays
// the hand-authored strong components from types/*.yaml, patches operations with
// a summary and a default 4xx response, derives the top-level tag list, prunes
// component schemas that nothing references, and writes the merged document.
//
// Usage:
//
//	go run ./tools/protobuf/openapi [spec-dir] [-o output] [-v]
//
// spec-dir defaults to protobuf/generated/openapi under the repo root; the
// output defaults to <spec-dir>/openapiv3-spec.yaml. This is the Go port of the
// former main.py and needs no Python/PyYAML.
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
	"time"

	"gopkg.in/yaml.v3"
)

// baseSpec is the merged document's skeleton — everything fixed rather than
// derived from the per-service specs. It is loaded from base.yaml in this
// directory (embedded so the tool stays self-contained); override it at runtime
// with -base.
//
//go:embed base.yaml
var baseSpec []byte

// badRequestResponse is the default 400 added to operations lacking any 4xx.
const badRequestResponse = `description: Bad Request
content:
  application/json:
    schema:
      $ref: '#/components/schemas/Status'
`

// httpMethods is the set of OpenAPI operation keys under a path item.
var httpMethods = map[string]bool{
	"get": true, "post": true, "put": true, "patch": true,
	"delete": true, "head": true, "options": true, "trace": true,
}

var verbose bool

func main() {
	root := repoRoot()
	defaultSpecDir := filepath.Join(root, "protobuf", "generated", "openapi")
	defaultOutput := filepath.Join(defaultSpecDir, "openapiv3-spec.yaml")

	var output string
	flag.StringVar(&output, "output", defaultOutput, "Path to write the merged spec")
	flag.StringVar(&output, "o", defaultOutput, "Path to write the merged spec (shorthand)")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose (debug) logging")
	flag.BoolVar(&verbose, "v", false, "Enable verbose (debug) logging (shorthand)")
	var basePath string
	flag.StringVar(&basePath, "base", "", "Path to the base spec YAML (default: embedded base.yaml)")
	flag.Parse()

	specDir := defaultSpecDir
	if flag.NArg() > 0 {
		specDir = flag.Arg(0)
	}
	specDir, _ = filepath.Abs(specDir)
	if fi, err := os.Stat(specDir); err != nil || !fi.IsDir() {
		fatalf("spec directory does not exist: %s", specDir)
	}

	base := baseSpec
	if basePath != "" {
		b, err := os.ReadFile(basePath)
		if err != nil {
			fatalf("read base spec %q: %v", basePath, err)
		}
		base = b
	}
	merged, err := parseYAML(string(base))
	if err != nil {
		fatalf("parse base spec: %v", err)
	}
	paths := mapGet(merged, "paths")
	paths.Style = 0 // base has `paths: {}` (flow); emit block once populated.
	components := mapGet(merged, "components")

	// Step 1: paths + weak components from *.openapi.yaml.
	mergePathsAndWeakComponents(specDir, paths, components)
	// Step 2: strong component definitions from types/*.yaml.
	mergeStrongComponents(specDir, components)
	// Step 3: patch operations with summary and 4xx responses.
	patchOperations(paths)
	// Step 4: populate the top-level tags list from operation tags.
	mapSet(merged, "tags", collectTags(paths))
	// Step 5: drop components nothing references.
	pruneUnusedComponents(merged)

	// Step 6: write the merged output.
	output, _ = filepath.Abs(output)
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		fatalf("create output dir: %v", err)
	}
	if err := writeYAML(output, merged); err != nil {
		fatalf("write output: %v", err)
	}

	tags := mapGet(merged, "tags")
	logf("INFO", "merged spec written to %s (%d paths, %d schemas, %d tags)",
		output, pairs(paths), pairs(schemasOf(components)), len(tags.Content))
}

// mergePathsAndWeakComponents loads every *.openapi.yaml (recursively) and folds
// its paths (last loaded wins) and components (first definition wins) in.
func mergePathsAndWeakComponents(specDir string, paths, components *yaml.Node) {
	files := findFiles(specDir, ".openapi.yaml")
	if len(files) == 0 {
		logf("WARN", "no *.openapi.yaml files found under %s", specDir)
	}
	for _, file := range files {
		doc, err := loadYAMLFile(file)
		if err != nil {
			logf("ERROR", "failed to parse %s: %v", file, err)
			continue
		}
		rel, _ := filepath.Rel(specDir, file)
		if doc == nil || doc.Kind != yaml.MappingNode {
			logf("WARN", "skipping non-object YAML: %s", rel)
			continue
		}
		logf("INFO", "merging %s", rel)

		if p := mapGet(doc, "paths"); p != nil && p.Kind == yaml.MappingNode {
			for i := 0; i+1 < len(p.Content); i += 2 {
				key := p.Content[i].Value
				if mapHas(paths, key) {
					logf("WARN", "duplicate path %q; keeping last loaded", key)
				}
				mapSet(paths, key, p.Content[i+1])
			}
		}
		if c := mapGet(doc, "components"); c != nil {
			mergeComponentSections(components, c, false)
		}
	}
}

// mergeStrongComponents overlays the hand-authored strong component definitions
// from types/*.yaml, overriding any weak definitions of the same name.
func mergeStrongComponents(specDir string, components *yaml.Node) {
	files, _ := filepath.Glob(filepath.Join(specDir, "types", "*.yaml"))
	sort.Strings(files)
	for _, file := range files {
		doc, err := loadYAMLFile(file)
		if err != nil {
			logf("ERROR", "failed to parse type file %s: %v", file, err)
			continue
		}
		if doc == nil || doc.Kind != yaml.MappingNode || mapGet(doc, "components") == nil {
			logf("DEBUG", "skipping non-component file: %s", filepath.Base(file))
			continue
		}
		logf("INFO", "loading strong components from %s", filepath.Base(file))
		mergeComponentSections(components, mapGet(doc, "components"), true)
	}
}

// mergeComponentSections merges each section of source into target. When a key
// already exists, preferStrong replaces it; otherwise the existing one is kept.
func mergeComponentSections(target, source *yaml.Node, preferStrong bool) {
	if source.Kind != yaml.MappingNode {
		return
	}
	for i := 0; i+1 < len(source.Content); i += 2 {
		section := source.Content[i].Value
		content := source.Content[i+1]
		if content.Kind != yaml.MappingNode {
			continue
		}
		targetSection := mapGet(target, section)
		if targetSection == nil {
			targetSection = newMapping()
			mapSet(target, section, targetSection)
		}
		for j := 0; j+1 < len(content.Content); j += 2 {
			key := content.Content[j].Value
			if mapHas(targetSection, key) {
				if preferStrong {
					logf("DEBUG", "overriding component %q in %q with strong type definition", key, section)
				} else {
					logf("DEBUG", "duplicate component %q in %q; keeping existing definition", key, section)
					continue
				}
			}
			mapSet(targetSection, key, content.Content[j+1])
		}
	}
}

// patchOperations gives every operation a summary (if missing), humanizes its
// tags, and adds a default 400 response when it declares no 4xx.
func patchOperations(paths *yaml.Node) {
	if paths.Kind != yaml.MappingNode {
		return
	}
	for i := 0; i+1 < len(paths.Content); i += 2 {
		pathKey := paths.Content[i].Value
		item := paths.Content[i+1]
		if item.Kind != yaml.MappingNode {
			continue
		}
		for j := 0; j+1 < len(item.Content); j += 2 {
			method := item.Content[j].Value
			op := item.Content[j+1]
			if !httpMethods[method] || op.Kind != yaml.MappingNode {
				continue
			}
			if tags := mapGet(op, "tags"); tags != nil && tags.Kind == yaml.SequenceNode {
				for _, t := range tags.Content {
					if t.Kind == yaml.ScalarNode {
						t.Value = splitCamelCase(t.Value)
					}
				}
			}
			if !mapHas(op, "summary") {
				summary := humanizeOperationID(scalar(mapGet(op, "operationId")))
				if summary == "" {
					summary = strings.ToUpper(method) + " " + pathKey
				}
				mapSet(op, "summary", strNode(summary))
			}
			responses := mapGet(op, "responses")
			if responses == nil {
				responses = newMapping()
				mapSet(op, "responses", responses)
			}
			if responses.Kind == yaml.MappingNode && !has4xx(responses) {
				if r, err := parseYAML(badRequestResponse); err == nil {
					mapSet(responses, "400", r)
				}
			}
		}
	}
}

// collectTags builds the top-level tags sequence from the tags used across all
// operations, de-duplicated and sorted.
func collectTags(paths *yaml.Node) *yaml.Node {
	names := map[string]bool{}
	if paths.Kind == yaml.MappingNode {
		for i := 0; i+1 < len(paths.Content); i += 2 {
			item := paths.Content[i+1]
			if item.Kind != yaml.MappingNode {
				continue
			}
			for j := 0; j+1 < len(item.Content); j += 2 {
				if !httpMethods[item.Content[j].Value] {
					continue
				}
				if tags := mapGet(item.Content[j+1], "tags"); tags != nil {
					for _, t := range tags.Content {
						if t.Kind == yaml.ScalarNode {
							names[t.Value] = true
						}
					}
				}
			}
		}
	}
	sorted := make([]string, 0, len(names))
	for n := range names {
		sorted = append(sorted, n)
	}
	sort.Strings(sorted)
	seq := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	for _, n := range sorted {
		m := newMapping()
		mapSet(m, "name", strNode(n))
		seq.Content = append(seq.Content, m)
	}
	return seq
}

// pruneUnusedComponents removes component schemas that nothing references,
// iterating because removing a schema can orphan others.
func pruneUnusedComponents(root *yaml.Node) {
	components := mapGet(root, "components")
	schemas := schemasOf(components)
	if schemas == nil || len(schemas.Content) == 0 {
		return
	}
	paths := mapGet(root, "paths")

	referenced := func() map[string]bool {
		refs := map[string]bool{}
		findRefs(paths, refs)
		findRefs(components, refs)
		names := map[string]bool{}
		for ref := range refs {
			if strings.HasPrefix(ref, "#/components/schemas/") {
				names[ref[strings.LastIndex(ref, "/")+1:]] = true
			}
		}
		return names
	}

	removed := 0
	for {
		used := referenced()
		var unused []string
		for i := 0; i+1 < len(schemas.Content); i += 2 {
			if name := schemas.Content[i].Value; !used[name] {
				unused = append(unused, name)
			}
		}
		if len(unused) == 0 {
			break
		}
		for _, name := range unused {
			deleteKey(schemas, name)
		}
		removed += len(unused)
	}
	if removed > 0 {
		logf("INFO", "pruned %d unreferenced component schema(s)", removed)
	}
}

// findRefs collects every $ref string reachable from n.
func findRefs(n *yaml.Node, refs map[string]bool) {
	if n == nil {
		return
	}
	if n.Kind == yaml.MappingNode {
		for i := 0; i+1 < len(n.Content); i += 2 {
			if n.Content[i].Value == "$ref" && n.Content[i+1].Kind == yaml.ScalarNode {
				refs[n.Content[i+1].Value] = true
			}
			findRefs(n.Content[i+1], refs)
		}
		return
	}
	for _, c := range n.Content { // sequence / document
		findRefs(c, refs)
	}
}

// --- text helpers ---

// splitCamelCase inserts a space at each lower/digit -> upper boundary, e.g.
// "BookingService" -> "Booking Service".
func splitCamelCase(name string) string {
	var b strings.Builder
	runes := []rune(name)
	for i, r := range runes {
		if i > 0 && r >= 'A' && r <= 'Z' {
			if p := runes[i-1]; (p >= 'a' && p <= 'z') || (p >= '0' && p <= '9') {
				b.WriteByte(' ')
			}
		}
		b.WriteRune(r)
	}
	return b.String()
}

// humanizeOperationID turns "ServiceName_MethodName" into "Method Name".
func humanizeOperationID(opID string) string {
	if opID == "" {
		return ""
	}
	parts := strings.SplitN(opID, "_", 2)
	return splitCamelCase(parts[len(parts)-1])
}

// --- yaml.Node helpers ---

func parseYAML(s string) (*yaml.Node, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(s), &doc); err != nil {
		return nil, err
	}
	return docRoot(&doc), nil
}

func loadYAMLFile(path string) (*yaml.Node, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	if len(doc.Content) == 0 {
		return nil, nil
	}
	return docRoot(&doc), nil
}

func writeYAML(path string, root *yaml.Node) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	if err := enc.Encode(root); err != nil {
		return err
	}
	return enc.Close()
}

func docRoot(n *yaml.Node) *yaml.Node {
	if n.Kind == yaml.DocumentNode && len(n.Content) > 0 {
		return n.Content[0]
	}
	return n
}

func mapGet(m *yaml.Node, key string) *yaml.Node {
	if m == nil || m.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return m.Content[i+1]
		}
	}
	return nil
}

func mapHas(m *yaml.Node, key string) bool { return mapGet(m, key) != nil }

// mapSet updates key in place when present, otherwise appends it (preserving
// insertion order, like the Python dict).
func mapSet(m *yaml.Node, key string, val *yaml.Node) {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			m.Content[i+1] = val
			return
		}
	}
	m.Content = append(m.Content, strNode(key), val)
}

func deleteKey(m *yaml.Node, key string) {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			m.Content = append(m.Content[:i], m.Content[i+2:]...)
			return
		}
	}
}

func strNode(s string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: s}
}

func newMapping() *yaml.Node {
	return &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
}

func scalar(n *yaml.Node) string {
	if n != nil && n.Kind == yaml.ScalarNode {
		return n.Value
	}
	return ""
}

func has4xx(responses *yaml.Node) bool {
	for i := 0; i+1 < len(responses.Content); i += 2 {
		if strings.HasPrefix(responses.Content[i].Value, "4") {
			return true
		}
	}
	return false
}

func schemasOf(components *yaml.Node) *yaml.Node {
	if components == nil {
		return nil
	}
	return mapGet(components, "schemas")
}

func pairs(m *yaml.Node) int {
	if m == nil {
		return 0
	}
	return len(m.Content) / 2
}

// --- misc helpers ---

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
