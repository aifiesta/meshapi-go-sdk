package meshapi

import (
	"encoding/json"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

type dumpedSurface struct {
	SchemaVersion string      `json:"schema_version"`
	SpecVersion   string      `json:"spec_version"`
	ModelFields   [][2]string `json:"model_fields"`
	Params        []string    `json:"params"`
}

func dumpSurface(t *testing.T) dumpedSurface {
	t.Helper()
	out, err := exec.Command("go", "run", "./cmd/dumpsurface").Output()
	if err != nil {
		t.Fatalf("go run ./cmd/dumpsurface: %v", err)
	}
	var s dumpedSurface
	if err := json.Unmarshal(out, &s); err != nil {
		t.Fatalf("decode surface: %v", err)
	}
	return s
}

func TestContractSurface_DumpContainsKnownMembers(t *testing.T) {
	s := dumpSurface(t)
	version, err := os.ReadFile(".meshapi-spec-version")
	if err != nil {
		t.Fatal(err)
	}
	if s.SchemaVersion != "1" || s.SpecVersion != strings.TrimSpace(string(version)) {
		t.Fatalf("unexpected envelope: %q / %q", s.SchemaVersion, s.SpecVersion)
	}
	fields := map[[2]string]bool{}
	for _, f := range s.ModelFields {
		fields[f] = true
	}
	for _, want := range [][2]string{
		{"ChatCompletionRequest", "model"},
		{"ChatCompletionRequest", "messages"},
		{"Message", "role"},
		{"ChatCompletionParams", "model"},
		{"ToolCall", "id"},
	} {
		if !fields[want] {
			t.Errorf("surface is missing field %v", want)
		}
	}
	params := map[string]bool{}
	for _, p := range s.Params {
		params[p] = true
	}
	for _, want := range []string{
		"POST /v1/chat/completions:model",
		"GET /v1/batches/{batch_id}:batch_id",
		"GET /v1/batches:after",
	} {
		if !params[want] {
			t.Errorf("surface is missing param %q", want)
		}
	}
}

func TestContractSurface_RegistryCoversPackage(t *testing.T) {
	registry, err := os.ReadFile("cmd/dumpsurface/registry.go")
	if err != nil {
		t.Fatal(err)
	}
	registered := map[string]bool{}
	for _, m := range regexp.MustCompile(`meshapi\.(\w+)\{\}`).FindAllStringSubmatch(string(registry), -1) {
		registered[m[1]] = true
	}
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				gd, ok := decl.(*ast.GenDecl)
				if !ok {
					continue
				}
				for _, spec := range gd.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					if _, isStruct := ts.Type.(*ast.StructType); isStruct && ts.Name.IsExported() && !registered[ts.Name.Name] {
						t.Errorf("struct %s is not in cmd/dumpsurface/registry.go", ts.Name.Name)
					}
				}
			}
		}
	}
}

// gatePython returns an interpreter that can import the gate, or skips the
// test locally; CI must have it installed.
func gatePython(t *testing.T) string {
	t.Helper()
	candidates := []string{os.Getenv("MESHAPI_CONTRACT_PYTHON"), "python", "python3"}
	for _, c := range candidates {
		if c == "" {
			continue
		}
		if exec.Command(c, "-c", "import meshapi_contract.gate").Run() == nil {
			return c
		}
	}
	if os.Getenv("CI") != "" {
		t.Fatal("meshapi_contract is not installed")
	}
	t.Skip("meshapi_contract not installed; set MESHAPI_CONTRACT_PYTHON")
	return ""
}

func runGate(t *testing.T, python, actual string) int {
	t.Helper()
	version, err := os.ReadFile(".meshapi-spec-version")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(python, "-m", "meshapi_contract.gate",
		"--spec", "contract/openapi.public.json",
		"--actual", actual,
		"--spec-version", strings.TrimSpace(string(version)),
		"--baseline", "contract-baseline.json")
	out, err := cmd.CombinedOutput()
	var exit *exec.ExitError
	switch {
	case err == nil:
		return 0
	case errors.As(err, &exit):
		t.Logf("gate output: %s", out)
		return exit.ExitCode()
	default:
		t.Fatalf("run gate: %v", err)
		return -1
	}
}

func writeSurface(t *testing.T, s dumpedSurface) string {
	t.Helper()
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "actual.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestContractSurface_GatePassesAndDriftFails(t *testing.T) {
	python := gatePython(t)
	s := dumpSurface(t)

	if code := runGate(t, python, writeSurface(t, s)); code != 0 {
		t.Fatalf("gate on the real surface exited %d, want 0", code)
	}

	// Dropping a covered field and a covered param must both be reported as drift.
	drop := [2]string{"ChatCompletionRequest", "model"}
	mutated := s
	mutated.ModelFields = nil
	for _, f := range s.ModelFields {
		if f != drop {
			mutated.ModelFields = append(mutated.ModelFields, f)
		}
	}
	if code := runGate(t, python, writeSurface(t, mutated)); code != 1 {
		t.Errorf("gate with a removed field exited %d, want 1", code)
	}

	mutated = s
	mutated.Params = nil
	for _, p := range s.Params {
		if p != "GET /v1/batches:limit" {
			mutated.Params = append(mutated.Params, p)
		}
	}
	if code := runGate(t, python, writeSurface(t, mutated)); code != 1 {
		t.Errorf("gate with a removed param exited %d, want 1", code)
	}
}
