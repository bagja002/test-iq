package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var defaultEnvCandidates = []string{
	".env",
	".env.local",
	"apps/api/.env",
	"apps/api/.env.local",
}

func LoadEnvFiles() ([]string, error) {
	loaded := make([]string, 0, len(defaultEnvCandidates))

	for _, candidate := range defaultEnvCandidates {
		path := filepath.Clean(candidate)
		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return loaded, fmt.Errorf("stat env file %s: %w", path, err)
		}

		if info.IsDir() {
			continue
		}

		if err := loadEnvFile(path); err != nil {
			return loaded, err
		}

		loaded = append(loaded, path)
	}

	return loaded, nil
}

func loadEnvFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open env file %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		key, value, ok, err := parseEnvLine(scanner.Text())
		if err != nil {
			return fmt.Errorf("%s:%d: %w", path, lineNumber, err)
		}
		if !ok {
			continue
		}

		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("set env %s from %s: %w", key, path, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read env file %s: %w", path, err)
	}

	return nil
}

func parseEnvLine(line string) (key string, value string, ok bool, err error) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return "", "", false, nil
	}

	if strings.HasPrefix(trimmed, "export ") {
		trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "export "))
	}

	parts := strings.SplitN(trimmed, "=", 2)
	if len(parts) != 2 {
		return "", "", false, fmt.Errorf("invalid env assignment")
	}

	key = strings.TrimSpace(parts[0])
	if key == "" {
		return "", "", false, fmt.Errorf("empty env key")
	}

	value = strings.TrimSpace(parts[1])
	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
			value = value[1 : len(value)-1]
		}
	}

	return key, value, true, nil
}
