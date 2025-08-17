package utils_test

import (
	"reflect"
	"testing"

	"github.com/next-trace/scg-logger/utils"
)

func TestSanitizeKV_EvenPairs(t *testing.T) {
	in := []any{"a", 1, "b", 2}
	out := utils.SanitizeKV(in)

	if !reflect.DeepEqual(in, out) {
		t.Fatalf("expected unchanged even kv, got %#v", out)
	}
}

func TestSanitizeKV_OddLengthAddsErrorAndDropsLast(t *testing.T) {
	in := []any{"a", 1, "b"}
	out := utils.SanitizeKV(in)

	if len(out)%2 != 0 {
		t.Fatalf("expected even length after sanitize, got %d", len(out))
	}

	foundErr := false

	for i := 0; i < len(out); i += 2 {
		if out[i] == "kv_error" && out[i+1] == "odd_length" {
			foundErr = true
			break
		}
	}

	if !foundErr {
		t.Fatalf("expected kv_error=odd_length in output, got %#v", out)
	}
}

func TestSanitizeKV_NonStringKeyBecomesEmpty(t *testing.T) {
	in := []any{123, "v"}
	out := utils.SanitizeKV(in)

	if len(out) != 2 {
		t.Fatalf("unexpected length: %d", len(out))
	}

	if s, ok := out[0].(string); !ok || s != "" {
		t.Fatalf("expected empty string key, got %#v", out[0])
	}
}
