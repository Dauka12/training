package config

import "testing"

func TestResolveListenAddr(t *testing.T) {
	t.Run("uses app addr when provided", func(t *testing.T) {
		addr := resolveListenAddr(":19090", "")
		if addr != ":19090" {
			t.Fatalf("expected explicit addr, got %q", addr)
		}
	})

	t.Run("builds addr from app port", func(t *testing.T) {
		addr := resolveListenAddr("", "18080")
		if addr != ":18080" {
			t.Fatalf("expected port-derived addr, got %q", addr)
		}
	})

	t.Run("falls back to default", func(t *testing.T) {
		addr := resolveListenAddr("", "")
		if addr != ":8080" {
			t.Fatalf("expected default addr, got %q", addr)
		}
	})
}
