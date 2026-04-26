package serverlesssdk

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Client struct {
	inputDir    string
	outputDir   string
	payloadPath string
	args        map[string]any
}

func New() (*Client, error) {
	inputDir := envOr("FAAS_INPUT_DIR", "/tmp/input")
	outputDir := envOr("FAAS_OUTPUT_DIR", "/tmp/output")
	payload := envOr("FAAS_PAYLOAD_PATH", "/tmp/payload.json")

	raw := envOr("FAAS_ARGS", "{}")

	args, err := parseArgs(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse FAAS_ARGS: %w", err)
	}

	return &Client{
		inputDir:    inputDir,
		outputDir:   outputDir,
		payloadPath: payload,
		args:        args,
	}, nil
}

// ReadPayload decodes the payload JSON into v.
func (c *Client) ReadPayload(v any) error {
	data, err := os.ReadFile(c.payloadPath)
	if err != nil {
		return fmt.Errorf("read payload: %w", err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("decode payload: %w", err)
	}
	return nil
}

// WriteJSON encodes v as JSON and writes it to stdout.
func (c *Client) WriteJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", " ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("write json: %w", err)
	}
	return nil
}

// GetFile returns the bytes of a file from the input directory.
func (c *Client) GetFile(name string) ([]byte, error) {
	data, err := os.ReadFile(filepath.Join(c.inputDir, name))
	if err != nil {
		return nil, fmt.Errorf("get file %q: %w", name, err)
	}
	return data, nil
}

// ListFiles returns all filenames in the input directory.
func (c *Client) ListFiles() ([]string, error) {
	entries, err := os.ReadDir(c.inputDir)
	if err != nil {
		return nil, fmt.Errorf("list files: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	return names, err
}

// WriteFile writes data to a file in the output directory.
func (c *Client) WriteFile(name string, data []byte) error {
	if err := os.MkdirAll(c.outputDir, 0o755); err != nil {
		return fmt.Errorf("ensure output directory: %w", err)
	}
	path := filepath.Join(c.outputDir, name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write file %q: %w", name, err)
	}
	return nil
}

// Args returns the parsed FAAS_ARGS map.
func (c *Client) Args() map[string]any { return c.args }

func envOr(key string, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

func parseArgs(raw string) (map[string]any, error) {
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return nil, err
	}
	return m, nil
}
