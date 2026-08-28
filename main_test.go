package main

import (
	"io"
	"os"
	"strings"
	"testing"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}

	os.Stdout = writer
	fn()

	if err := writer.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}
	os.Stdout = originalStdout

	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read captured stdout: %v", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("close stdout reader: %v", err)
	}

	return string(output)
}

func TestGetRandomApiKeyDoesNotLogKeyMaterial(t *testing.T) {
	testCases := []struct {
		name string
		key  string
	}{
		{name: "long key", key: "sk-super-secret-completion-key-1234567890"},
		{name: "short key", key: "short-key"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var selectedKey string
			output := captureStdout(t, func() {
				selectedKey = getRandomApiKey(testCase.key)
			})

			if selectedKey != testCase.key {
				t.Fatalf("selected key = %q, want %q", selectedKey, testCase.key)
			}

			fragments := []string{
				testCase.key,
				testCase.key[:4],
				testCase.key[len(testCase.key)-4:],
			}
			for _, fragment := range fragments {
				if strings.Contains(output, fragment) {
					t.Errorf("stdout leaked API key material %q: %q", fragment, output)
				}
			}

			if !strings.Contains(output, "Code completion API Key index:") {
				t.Errorf("stdout did not include the selected key index: %q", output)
			}
		})
	}
}
