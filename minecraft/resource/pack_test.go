package resource

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReadURLContextCanceled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := ReadURLContext(ctx, "http://127.0.0.1/resource.mcpack"); !errors.Is(err, context.Canceled) {
		t.Fatalf("ReadURLContext error = %v, want context canceled", err)
	}
}

func TestReadURLContextLimitRejectsOversizedBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("12345"))
	}))
	defer server.Close()

	if _, err := ReadURLContextLimit(context.Background(), server.URL, 4); err == nil || !strings.Contains(err.Error(), "exceeds limit") {
		t.Fatalf("ReadURLContextLimit error = %v, want exceeds limit", err)
	}
}

func TestReadURLContextLimitDoesNotUnwrapNestedURLArchive(t *testing.T) {
	t.Parallel()

	inner := testPackArchive(t)
	outer := new(bytes.Buffer)
	zw := zip.NewWriter(outer)
	w, err := zw.Create("pack.zip")
	if err != nil {
		t.Fatalf("create nested zip entry: %v", err)
	}
	if _, err := w.Write(inner); err != nil {
		t.Fatalf("write nested zip entry: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close outer zip: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(outer.Bytes())
	}))
	defer server.Close()

	if _, err := ReadURLContextLimit(context.Background(), server.URL, uint64(outer.Len())); err == nil {
		t.Fatal("ReadURLContextLimit succeeded for nested URL archive, want error")
	}
}

func testPackArchive(t *testing.T) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	w, err := zw.Create("manifest.json")
	if err != nil {
		t.Fatalf("create manifest: %v", err)
	}
	_, _ = w.Write([]byte(`{
		"format_version": 2,
		"header": {
			"name": "test pack",
			"description": "test pack",
			"uuid": "550e8400-e29b-41d4-a716-446655440000",
			"version": [1, 0, 0],
			"min_engine_version": [1, 20, 0]
		},
		"modules": [{
			"description": "test pack",
			"type": "resources",
			"uuid": "550e8400-e29b-41d4-a716-446655440001",
			"version": [1, 0, 0]
		}]
	}`))
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
}
