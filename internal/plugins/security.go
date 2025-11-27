package plugins

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// SecurityConfig holds plugin security settings
type SecurityConfig struct {
	// RequireVerification enforces checksum verification for all plugins
	RequireVerification bool
	// AllowlistPath is the path to the plugin allowlist file
	AllowlistPath string
	// AllowUnsignedPlugins allows loading plugins without checksum verification (with warning)
	AllowUnsignedPlugins bool
	// TrustedDirectories is a list of directories considered safe for plugin loading
	TrustedDirectories []string
}

// DefaultSecurityConfig returns secure default settings
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		RequireVerification:  false, // Default to false for backward compatibility
		AllowlistPath:        "",
		AllowUnsignedPlugins: true, // Default to true for backward compatibility
		TrustedDirectories:   []string{},
	}
}

// PluginSecurityError represents a security-related error during plugin operations
type PluginSecurityError struct {
	Plugin    string
	Operation string
	Reason    string
	Cause     error
}

func (e *PluginSecurityError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("security error for plugin %s during %s: %s (%v)", e.Plugin, e.Operation, e.Reason, e.Cause)
	}
	return fmt.Sprintf("security error for plugin %s during %s: %s", e.Plugin, e.Operation, e.Reason)
}

func (e *PluginSecurityError) Unwrap() error {
	return e.Cause
}

// PathTraversalError indicates a path traversal attempt was detected
type PathTraversalError struct {
	Path   string
	Reason string
}

func (e *PathTraversalError) Error() string {
	return fmt.Sprintf("path traversal detected: %s - %s", e.Path, e.Reason)
}

// AllowlistEntry represents a single entry in the plugin allowlist
type AllowlistEntry struct {
	Name     string
	Checksum string
	Enabled  bool
}

// PluginAllowlist manages the list of allowed plugins with their checksums
type PluginAllowlist struct {
	entries map[string]AllowlistEntry
	mu      sync.RWMutex
}

// NewPluginAllowlist creates a new empty allowlist
func NewPluginAllowlist() *PluginAllowlist {
	return &PluginAllowlist{
		entries: make(map[string]AllowlistEntry),
	}
}

// LoadAllowlistFromFile loads an allowlist from a file
// File format: one entry per line, format: "plugin_name:sha256_checksum" or "plugin_name:sha256_checksum:disabled"
func LoadAllowlistFromFile(path string) (*PluginAllowlist, error) {
	if path == "" {
		return NewPluginAllowlist(), nil
	}

	// Validate path for traversal sequences before opening
	if containsPathTraversal(path) {
		return nil, &PathTraversalError{
			Path:   path,
			Reason: "allowlist file path contains traversal sequences",
		}
	}

	file, err := os.Open(path) // #nosec G304 -- path validated above for traversal
	if err != nil {
		if os.IsNotExist(err) {
			return NewPluginAllowlist(), nil
		}
		return nil, fmt.Errorf("failed to open allowlist file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	allowlist := NewPluginAllowlist()
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid allowlist entry at line %d: expected 'name:checksum' format", lineNum)
		}

		name := strings.TrimSpace(parts[0])
		checksum := strings.TrimSpace(parts[1])

		if name == "" || checksum == "" {
			return nil, fmt.Errorf("invalid allowlist entry at line %d: name and checksum cannot be empty", lineNum)
		}

		// Validate checksum format (SHA256 = 64 hex characters)
		if len(checksum) != 64 {
			return nil, fmt.Errorf("invalid checksum at line %d: expected 64 character SHA256 hash", lineNum)
		}

		if _, err := hex.DecodeString(checksum); err != nil {
			return nil, fmt.Errorf("invalid checksum at line %d: not valid hex encoding", lineNum)
		}

		enabled := true
		if len(parts) >= 3 && strings.TrimSpace(parts[2]) == "disabled" {
			enabled = false
		}

		allowlist.entries[name] = AllowlistEntry{
			Name:     name,
			Checksum: strings.ToLower(checksum),
			Enabled:  enabled,
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading allowlist file: %w", err)
	}

	return allowlist, nil
}

// IsAllowed checks if a plugin with the given name and checksum is allowed
func (a *PluginAllowlist) IsAllowed(name, checksum string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	entry, exists := a.entries[name]
	if !exists {
		return false
	}

	return entry.Enabled && strings.EqualFold(entry.Checksum, checksum)
}

// HasEntry checks if a plugin name exists in the allowlist (regardless of checksum)
func (a *PluginAllowlist) HasEntry(name string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	_, exists := a.entries[name]
	return exists
}

// GetExpectedChecksum returns the expected checksum for a plugin
func (a *PluginAllowlist) GetExpectedChecksum(name string) (string, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	entry, exists := a.entries[name]
	if !exists {
		return "", false
	}
	return entry.Checksum, true
}

// IsEmpty returns true if the allowlist has no entries
func (a *PluginAllowlist) IsEmpty() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.entries) == 0
}

// CalculateFileChecksum computes the SHA256 checksum of a file
func CalculateFileChecksum(path string) (string, error) {
	// Defense-in-depth: validate path even though callers may do their own checks
	if containsPathTraversal(path) {
		return "", &PathTraversalError{
			Path:   path,
			Reason: "file path contains traversal sequences",
		}
	}

	file, err := os.Open(path) // #nosec G304 -- path validated above for traversal
	if err != nil {
		return "", fmt.Errorf("failed to open file for checksum: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to compute checksum: %w", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// ValidatePluginPath validates a plugin directory path for security issues
func ValidatePluginPath(pluginDir string) (string, error) {
	if pluginDir == "" {
		return "", &PathTraversalError{
			Path:   pluginDir,
			Reason: "plugin directory path cannot be empty",
		}
	}

	// Check for obvious path traversal patterns before path resolution
	if containsPathTraversal(pluginDir) {
		return "", &PathTraversalError{
			Path:   pluginDir,
			Reason: "path contains traversal sequences (..)",
		}
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(pluginDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve plugin directory path: %w", err)
	}

	// Clean the path to remove any redundant separators or dots
	cleanPath := filepath.Clean(absPath)

	// Note: cleanPath may differ from absPath after filepath.Clean removes
	// redundant separators or dots. This is expected and safe since we already
	// checked for explicit path traversal sequences above.

	return cleanPath, nil
}

// containsPathTraversal checks for path traversal sequences in the original input
func containsPathTraversal(path string) bool {
	// Check for various path traversal patterns
	patterns := []string{
		"..",
		"..\\",
		"../",
	}

	normalizedPath := filepath.ToSlash(path)

	for _, pattern := range patterns {
		if strings.Contains(normalizedPath, pattern) {
			return true
		}
	}

	return false
}

// IsPathInTrustedDirectory checks if a path is within any of the trusted directories
func IsPathInTrustedDirectory(path string, trustedDirs []string) bool {
	if len(trustedDirs) == 0 {
		return false
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	for _, trusted := range trustedDirs {
		absTrusted, err := filepath.Abs(trusted)
		if err != nil {
			continue
		}

		// Check if path is within trusted directory
		rel, err := filepath.Rel(absTrusted, absPath)
		if err != nil {
			continue
		}

		// If the relative path doesn't start with "..", it's within the trusted dir
		if !strings.HasPrefix(rel, "..") {
			return true
		}
	}

	return false
}

// IsPathInWorkingDirectory checks if a path is within the current working directory
func IsPathInWorkingDirectory(path string) (bool, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return false, fmt.Errorf("failed to get current working directory: %w", err)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return false, fmt.Errorf("failed to resolve path: %w", err)
	}

	rel, err := filepath.Rel(cwd, absPath)
	if err != nil {
		return false, nil
	}

	// If relative path starts with "..", it's outside cwd
	return !strings.HasPrefix(rel, ".."), nil
}

// PluginLoadEvent represents a plugin load attempt for logging
type PluginLoadEvent struct {
	Timestamp       time.Time
	PluginPath      string
	PluginName      string
	Checksum        string
	Success         bool
	Error           string
	SecurityWarning string
}

// PluginSecurityLogger handles logging of plugin security events
type PluginSecurityLogger struct {
	events []PluginLoadEvent
	mu     sync.Mutex
}

// NewPluginSecurityLogger creates a new security logger
func NewPluginSecurityLogger() *PluginSecurityLogger {
	return &PluginSecurityLogger{
		events: make([]PluginLoadEvent, 0),
	}
}

// LogLoadAttempt records a plugin load attempt
func (l *PluginSecurityLogger) LogLoadAttempt(event PluginLoadEvent) {
	l.mu.Lock()
	defer l.mu.Unlock()

	event.Timestamp = time.Now()
	l.events = append(l.events, event)

	// Print to stderr for visibility
	if event.SecurityWarning != "" {
		fmt.Fprintf(os.Stderr, "[SECURITY WARNING] Plugin %s: %s\n", event.PluginPath, event.SecurityWarning)
	}

	if event.Success {
		fmt.Fprintf(os.Stderr, "[PLUGIN] Loaded: %s (checksum: %s)\n", event.PluginPath, truncateChecksum(event.Checksum))
	} else if event.Error != "" {
		fmt.Fprintf(os.Stderr, "[PLUGIN] Failed to load %s: %s\n", event.PluginPath, event.Error)
	}
}

// GetEvents returns all logged events
func (l *PluginSecurityLogger) GetEvents() []PluginLoadEvent {
	l.mu.Lock()
	defer l.mu.Unlock()

	result := make([]PluginLoadEvent, len(l.events))
	copy(result, l.events)
	return result
}

// truncateChecksum returns the first 12 characters of a checksum for display
func truncateChecksum(checksum string) string {
	if len(checksum) > 12 {
		return checksum[:12] + "..."
	}
	return checksum
}

// VerifyPlugin performs security verification on a plugin before loading
func VerifyPlugin(path string, config *SecurityConfig, allowlist *PluginAllowlist) (*PluginLoadEvent, error) {
	event := &PluginLoadEvent{
		PluginPath: path,
		PluginName: filepath.Base(path),
	}

	// Calculate checksum
	checksum, err := CalculateFileChecksum(path)
	if err != nil {
		event.Error = fmt.Sprintf("failed to calculate checksum: %v", err)
		return event, &PluginSecurityError{
			Plugin:    path,
			Operation: "checksum",
			Reason:    "failed to calculate file checksum",
			Cause:     err,
		}
	}
	event.Checksum = checksum

	// Check path safety
	inWorkDir, err := IsPathInWorkingDirectory(path)
	if err != nil {
		event.SecurityWarning = fmt.Sprintf("could not verify path safety: %v", err)
	} else if !inWorkDir {
		inTrusted := IsPathInTrustedDirectory(path, config.TrustedDirectories)
		if !inTrusted {
			event.SecurityWarning = "plugin loaded from directory outside current working directory and trusted paths"
		}
	}

	// If allowlist is provided and not empty, verify against it
	if allowlist != nil && !allowlist.IsEmpty() {
		pluginName := filepath.Base(path)

		if !allowlist.IsAllowed(pluginName, checksum) {
			if allowlist.HasEntry(pluginName) {
				expectedChecksum, _ := allowlist.GetExpectedChecksum(pluginName)
				event.Error = fmt.Sprintf("checksum mismatch: expected %s, got %s", truncateChecksum(expectedChecksum), truncateChecksum(checksum))
				return event, &PluginSecurityError{
					Plugin:    path,
					Operation: "verification",
					Reason:    "plugin checksum does not match allowlist",
				}
			}
			event.Error = "plugin not in allowlist"
			return event, &PluginSecurityError{
				Plugin:    path,
				Operation: "verification",
				Reason:    "plugin not found in allowlist",
			}
		}
	} else if config.RequireVerification {
		// Verification required but no allowlist provided
		event.Error = "verification required but no allowlist configured"
		return event, &PluginSecurityError{
			Plugin:    path,
			Operation: "verification",
			Reason:    "plugin verification required but no allowlist is configured",
		}
	} else if !config.AllowUnsignedPlugins {
		event.Error = "unsigned plugins not allowed"
		return event, &PluginSecurityError{
			Plugin:    path,
			Operation: "verification",
			Reason:    "unsigned plugins are not allowed by security policy",
		}
	} else {
		// No verification, add warning
		event.SecurityWarning = "plugin loaded without checksum verification"
	}

	event.Success = true
	return event, nil
}
