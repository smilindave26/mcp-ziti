package tools

import (
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestStripDenylistKeys_FlatObject(t *testing.T) {
	input := map[string]any{
		"id":      "abc123",
		"name":    "test",
		"_links":  map[string]any{"self": "http://example.com"},
		"envInfo": map[string]any{"os": "linux"},
		"sdkInfo": map[string]any{"version": "1.0"},
	}
	result := stripDenylistKeys(input).(map[string]any)

	if _, ok := result["_links"]; ok {
		t.Error("expected _links to be stripped")
	}
	if _, ok := result["envInfo"]; ok {
		t.Error("expected envInfo to be stripped")
	}
	if _, ok := result["sdkInfo"]; ok {
		t.Error("expected sdkInfo to be stripped")
	}
	if result["id"] != "abc123" {
		t.Error("expected id to be preserved")
	}
	if result["name"] != "test" {
		t.Error("expected name to be preserved")
	}
}

func TestStripDenylistKeys_Array(t *testing.T) {
	input := []any{
		map[string]any{"id": "1", "_links": "x"},
		map[string]any{"id": "2", "envInfo": "y"},
	}
	result := stripDenylistKeys(input).([]any)

	for i, item := range result {
		m := item.(map[string]any)
		if _, ok := m["_links"]; ok {
			t.Errorf("item %d: expected _links to be stripped", i)
		}
		if _, ok := m["envInfo"]; ok {
			t.Errorf("item %d: expected envInfo to be stripped", i)
		}
		if m["id"] == nil {
			t.Errorf("item %d: expected id to be preserved", i)
		}
	}
}

func TestStripDenylistKeys_Nested(t *testing.T) {
	input := map[string]any{
		"data": map[string]any{
			"id":     "nested",
			"_links": "remove-me",
			"child": map[string]any{
				"sdkInfo": "also-remove",
				"value":   42.0,
			},
		},
	}
	result := stripDenylistKeys(input).(map[string]any)
	data := result["data"].(map[string]any)

	if _, ok := data["_links"]; ok {
		t.Error("expected nested _links to be stripped")
	}
	child := data["child"].(map[string]any)
	if _, ok := child["sdkInfo"]; ok {
		t.Error("expected deeply nested sdkInfo to be stripped")
	}
	if child["value"] != 42.0 {
		t.Error("expected value to be preserved")
	}
}

func TestStripDenylistKeys_SimpleMap(t *testing.T) {
	input := map[string]any{"status": "deleted", "id": "xyz"}
	result := stripDenylistKeys(input).(map[string]any)

	if result["status"] != "deleted" || result["id"] != "xyz" {
		t.Error("simple map should pass through unchanged")
	}
}

func TestStripDenylistKeys_NilAndScalar(t *testing.T) {
	if stripDenylistKeys(nil) != nil {
		t.Error("nil should return nil")
	}
	if stripDenylistKeys("hello") != "hello" {
		t.Error("string scalar should pass through")
	}
	if stripDenylistKeys(42.0) != 42.0 {
		t.Error("number scalar should pass through")
	}
}

func TestJsonResult_StripsFields(t *testing.T) {
	input := struct {
		ID   string         `json:"id"`
		Name string         `json:"name"`
		Links map[string]any `json:"_links"`
	}{
		ID:   "test-id",
		Name: "test-name",
		Links: map[string]any{"self": "http://example.com"},
	}

	result, out, err := jsonResult(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Error("expected nil Out")
	}
	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content block, got %d", len(result.Content))
	}

	text := result.Content[0].(*mcp.TextContent).Text
	var parsed map[string]any
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("failed to parse result JSON: %v", err)
	}

	if _, ok := parsed["_links"]; ok {
		t.Error("expected _links to be stripped from jsonResult output")
	}
	if parsed["id"] != "test-id" {
		t.Error("expected id to be preserved in jsonResult output")
	}
	if parsed["name"] != "test-name" {
		t.Error("expected name to be preserved in jsonResult output")
	}
}
