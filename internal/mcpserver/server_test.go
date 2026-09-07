package mcpserver

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// connectClient wires an in-memory client/server pair — a genuine MCP
// protocol round-trip (initialize handshake, tools/list, tools/call) with
// no subprocess or stdio involved, unlike a hand-called Go function.
func connectClient(t *testing.T) *mcp.ClientSession {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	server := NewServer()
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}
	t.Cleanup(func() { _ = serverSession.Close() })

	client := mcp.NewClient(&mcp.Implementation{Name: "genitz-test-client", Version: "0.0.0"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { _ = clientSession.Close() })

	return clientSession
}

func callTool(t *testing.T, cs *mcp.ClientSession, name string, args any) *mcp.CallToolResult {
	t.Helper()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("CallTool(%s): %v", name, err)
	}
	return res
}

func decodeStructured(t *testing.T, res *mcp.CallToolResult, out any) {
	t.Helper()
	raw, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatalf("marshal structured content: %v", err)
	}
	if err := json.Unmarshal(raw, out); err != nil {
		t.Fatalf("unmarshal structured content: %v", err)
	}
}

func TestListTools(t *testing.T) {
	cs := connectClient(t)

	toolsRes, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	want := map[string]bool{
		"search_dependencies":         false,
		"get_dependency_info":         false,
		"list_presets":                false,
		"list_installed_dependencies": false,
		"audit_project":               false,
		"add_dependencies":            false,
		"remove_dependencies":         false,
		"scaffold_project":            false,
	}
	for _, tool := range toolsRes.Tools {
		if _, ok := want[tool.Name]; ok {
			want[tool.Name] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("expected tool %q to be registered, but ListTools did not return it", name)
		}
	}
}

func TestSearchDependencies(t *testing.T) {
	cs := connectClient(t)

	res := callTool(t, cs, "search_dependencies", map[string]any{"query": "redis"})
	if res.IsError {
		t.Fatalf("search_dependencies returned an error result: %+v", res.Content)
	}

	var out searchResult
	decodeStructured(t, res, &out)
	if len(out.Matches) == 0 {
		t.Fatalf("expected at least one match for %q, got none", "redis")
	}
	found := false
	for _, m := range out.Matches {
		if m.ID == "redis" {
			found = true
			if m.DocsURL == "" {
				t.Errorf("expected DocsURL to be populated for redis match")
			}
		}
	}
	if !found {
		t.Errorf("expected a match with ID \"redis\", got %+v", out.Matches)
	}
}

func TestGetDependencyInfoUnknownID(t *testing.T) {
	cs := connectClient(t)

	res := callTool(t, cs, "get_dependency_info", map[string]any{"id": "not-a-real-dependency-id"})
	if !res.IsError {
		t.Fatalf("expected get_dependency_info to report an error for an unknown ID, got %+v", res)
	}
}

func TestListPresets(t *testing.T) {
	cs := connectClient(t)

	res := callTool(t, cs, "list_presets", map[string]any{})
	if res.IsError {
		t.Fatalf("list_presets returned an error result: %+v", res.Content)
	}

	var out listPresetsResult
	decodeStructured(t, res, &out)
	if len(out.Presets) == 0 {
		t.Fatalf("expected at least one built-in preset, got none")
	}
}

func TestResolveMCPPathNoRootSetAllowsAnyPath(t *testing.T) {
	got, err := resolveMCPPath("/some/arbitrary/path")
	if err != nil {
		t.Fatalf("expected no error with GENITZ_MCP_ROOT unset, got: %v", err)
	}
	if got != "/some/arbitrary/path" {
		t.Errorf("expected resolved path to be unchanged absolute input, got %q", got)
	}
}

func TestResolveMCPPathWithinRootSucceeds(t *testing.T) {
	root := t.TempDir()
	t.Setenv("GENITZ_MCP_ROOT", root)

	inside := filepath.Join(root, "sub", "project")
	got, err := resolveMCPPath(inside)
	if err != nil {
		t.Fatalf("expected path inside root to be allowed, got error: %v", err)
	}
	if got != inside {
		t.Errorf("expected resolved path %q, got %q", inside, got)
	}
}

func TestResolveMCPPathEscapingRootIsRejected(t *testing.T) {
	root := t.TempDir()
	t.Setenv("GENITZ_MCP_ROOT", root)

	cases := []string{
		filepath.Join(root, "..", "escaped"),
		"/etc",
		filepath.Dir(root), // parent of root
	}
	for _, raw := range cases {
		if _, err := resolveMCPPath(raw); err == nil {
			t.Errorf("expected resolveMCPPath(%q) to reject a path escaping root %q, got no error", raw, root)
		}
	}
}

func TestListInstalledDependenciesRejectsPathEscapingRoot(t *testing.T) {
	root := t.TempDir()
	t.Setenv("GENITZ_MCP_ROOT", root)

	cs := connectClient(t)
	res := callTool(t, cs, "list_installed_dependencies", map[string]any{"dir": "/etc"})
	if !res.IsError {
		t.Fatalf("expected list_installed_dependencies to reject a dir escaping GENITZ_MCP_ROOT, got success: %+v", res)
	}
}
