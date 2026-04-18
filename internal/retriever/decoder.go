package retriever

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

// DecodeSecrets decodes base64-encoded secrets from artifact content
func DecodeSecrets(artifactZip []byte) (map[string]string, error) {
	// Open the zip archive
	reader, err := zip.NewReader(bytes.NewReader(artifactZip), int64(len(artifactZip)))
	if err != nil {
		return nil, fmt.Errorf("failed to open artifact zip: %w", err)
	}

	// Find and read the secrets.txt file
	var secretsContent []byte
	for _, file := range reader.File {
		if file.Name == "secrets.txt" || strings.HasSuffix(file.Name, "/secrets.txt") {
			rc, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open secrets file: %w", err)
			}
			defer rc.Close()

			secretsContent, err = io.ReadAll(rc)
			if err != nil {
				return nil, fmt.Errorf("failed to read secrets file: %w", err)
			}
			break
		}
	}

	if secretsContent == nil {
		return nil, fmt.Errorf("secrets.txt not found in artifact")
	}

	// Parse the secrets
	secrets := make(map[string]string)
	lines := strings.Split(string(secretsContent), "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		name := parts[0]
		encodedValue := parts[1]

		if encodedValue == "__NOT_FOUND__" {
			return nil, fmt.Errorf("secret %s not found or not accessible", name)
		}

		// Decode base64
		decodedValue, err := base64.StdEncoding.DecodeString(encodedValue)
		if err != nil {
			return nil, fmt.Errorf("failed to decode secret %s: %w", name, err)
		}

		secrets[name] = string(decodedValue)
	}

	return secrets, nil
}
