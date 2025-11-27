package plugins

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultSecurityConfig(t *testing.T) {
	config := DefaultSecurityConfig()

	if config.RequireVerification {
		t.Error("RequireVerification should be false by default")
	}
	if config.AllowlistPath != "" {
		t.Error("AllowlistPath should be empty by default")
	}
	if !config.AllowUnsignedPlugins {
		t.Error("AllowUnsignedPlugins should be true by default")
	}
	if len(config.TrustedDirectories) != 0 {
		t.Error("TrustedDirectories should be empty by default")
	}
}

func TestPluginSecurityError(t *testing.T) {
	err := &PluginSecurityError{
		Plugin:    "test-plugin.so",
		Operation: "verification",
		Reason:    "checksum mismatch",
	}

	expected := "security error for plugin test-plugin.so during verification: checksum mismatch"
	if err.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, err.Error())
	}

	// Test with cause
	cause := &PathTraversalError{Path: "/etc/evil", Reason: "outside allowed paths"}
	errWithCause := &PluginSecurityError{
		Plugin:    "evil.so",
		Operation: "path-check",
		Reason:    "invalid path",
		Cause:     cause,
	}

	if errWithCause.Unwrap() != cause {
		t.Error("Unwrap should return the cause")
	}
}

func TestPathTraversalError(t *testing.T) {
	err := &PathTraversalError{
		Path:   "../../../etc/passwd",
		Reason: "contains traversal sequences",
	}

	expected := "path traversal detected: ../../../etc/passwd - contains traversal sequences"
	if err.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, err.Error())
	}
}

func TestNewPluginAllowlist(t *testing.T) {
	allowlist := NewPluginAllowlist()

	if allowlist == nil {
		t.Fatal("NewPluginAllowlist should not return nil")
	}

	if !allowlist.IsEmpty() {
		t.Error("New allowlist should be empty")
	}
}

func TestLoadAllowlistFromFile(t *testing.T) {
	// Test with empty path
	allowlist, err := LoadAllowlistFromFile("")
	if err != nil {
		t.Errorf("Empty path should not error: %v", err)
	}
	if !allowlist.IsEmpty() {
		t.Error("Allowlist with empty path should be empty")
	}

	// Test with non-existent file
	allowlist, err = LoadAllowlistFromFile("/nonexistent/path/allowlist.txt")
	if err != nil {
		t.Errorf("Non-existent file should not error: %v", err)
	}
	if !allowlist.IsEmpty() {
		t.Error("Allowlist with non-existent file should be empty")
	}

	// Create temp file with valid content
	tmpDir := t.TempDir()
	allowlistPath := filepath.Join(tmpDir, "allowlist.txt")
	content := `# Comment line
myplugin.so:a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3
otherplugin.so:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855:disabled
`
	err = os.WriteFile(allowlistPath, []byte(content), 0600)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	allowlist, err = LoadAllowlistFromFile(allowlistPath)
	if err != nil {
		t.Errorf("Valid allowlist file should not error: %v", err)
	}

	if allowlist.IsEmpty() {
		t.Error("Loaded allowlist should not be empty")
	}

	// Check myplugin.so is allowed
	if !allowlist.IsAllowed("myplugin.so", "a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3") {
		t.Error("myplugin.so with correct checksum should be allowed")
	}

	// Check wrong checksum is not allowed
	if allowlist.IsAllowed("myplugin.so", "wrongchecksum1234567890123456789012345678901234567890123456") {
		t.Error("myplugin.so with wrong checksum should not be allowed")
	}

	// Check disabled plugin is not allowed
	if allowlist.IsAllowed("otherplugin.so", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855") {
		t.Error("Disabled plugin should not be allowed")
	}

	// Check HasEntry for disabled plugin
	if !allowlist.HasEntry("otherplugin.so") {
		t.Error("HasEntry should return true for disabled plugin")
	}
}

func TestLoadAllowlistFromFile_InvalidContent(t *testing.T) {
	tmpDir := t.TempDir()

	// Test invalid format (missing colon)
	invalidPath := filepath.Join(tmpDir, "invalid.txt")
	err := os.WriteFile(invalidPath, []byte("invalid_line_without_colon"), 0600)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err = LoadAllowlistFromFile(invalidPath)
	if err == nil {
		t.Error("Should error on invalid format")
	}

	// Test invalid checksum length
	invalidChecksumPath := filepath.Join(tmpDir, "invalid_checksum.txt")
	err = os.WriteFile(invalidChecksumPath, []byte("plugin.so:shortchecksum"), 0600)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err = LoadAllowlistFromFile(invalidChecksumPath)
	if err == nil {
		t.Error("Should error on invalid checksum length")
	}

	// Test invalid hex encoding
	invalidHexPath := filepath.Join(tmpDir, "invalid_hex.txt")
	err = os.WriteFile(invalidHexPath, []byte("plugin.so:gggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggg"), 0600)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err = LoadAllowlistFromFile(invalidHexPath)
	if err == nil {
		t.Error("Should error on invalid hex encoding")
	}
}

func TestAllowlistGetExpectedChecksum(t *testing.T) {
	allowlist := NewPluginAllowlist()
	allowlist.entries["test.so"] = AllowlistEntry{
		Name:     "test.so",
		Checksum: "a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3",
		Enabled:  true,
	}

	checksum, found := allowlist.GetExpectedChecksum("test.so")
	if !found {
		t.Error("Should find entry for test.so")
	}
	if checksum != "a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3" {
		t.Errorf("Unexpected checksum: %s", checksum)
	}

	_, found = allowlist.GetExpectedChecksum("nonexistent.so")
	if found {
		t.Error("Should not find entry for nonexistent.so")
	}
}

func TestCalculateFileChecksum(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Empty file should have known SHA256 hash
	err := os.WriteFile(testFile, []byte{}, 0600)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	checksum, err := CalculateFileChecksum(testFile)
	if err != nil {
		t.Errorf("Should not error: %v", err)
	}

	// SHA256 of empty file
	expected := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if checksum != expected {
		t.Errorf("Expected %s, got %s", expected, checksum)
	}

	// Test with content
	err = os.WriteFile(testFile, []byte("hello"), 0600)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	checksum, err = CalculateFileChecksum(testFile)
	if err != nil {
		t.Errorf("Should not error: %v", err)
	}

	// SHA256 of "hello"
	expected = "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if checksum != expected {
		t.Errorf("Expected %s, got %s", expected, checksum)
	}

	// Test non-existent file
	_, err = CalculateFileChecksum("/nonexistent/file.txt")
	if err == nil {
		t.Error("Should error on non-existent file")
	}
}

func TestValidatePluginPath(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		shouldErr bool
	}{
		{"empty path", "", true},
		{"path with traversal", "../../../etc/evil", true},
		{"path with backslash traversal", "..\\..\\evil", true},
		{"valid relative path", "./plugins", false},
		{"valid absolute path", "/tmp/plugins", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidatePluginPath(tt.path)
			if tt.shouldErr && err == nil {
				t.Errorf("Expected error for path %q", tt.path)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error for path %q: %v", tt.path, err)
			}
		})
	}
}

func TestContainsPathTraversal(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"../evil", true},
		{"..\\evil", true},
		{"./safe", false},
		{"/absolute/path", false},
		{"relative/path", false},
		{"path/with/../traversal", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := containsPathTraversal(tt.path)
			if result != tt.expected {
				t.Errorf("containsPathTraversal(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestIsPathInTrustedDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	trustedDir := filepath.Join(tmpDir, "trusted")
	err := os.MkdirAll(trustedDir, 0750)
	if err != nil {
		t.Fatalf("Failed to create trusted dir: %v", err)
	}

	untrustedDir := filepath.Join(tmpDir, "untrusted")
	err = os.MkdirAll(untrustedDir, 0750)
	if err != nil {
		t.Fatalf("Failed to create untrusted dir: %v", err)
	}

	// Empty trusted directories
	if IsPathInTrustedDirectory(trustedDir, nil) {
		t.Error("Should return false when no trusted directories")
	}

	// Path in trusted directory
	trustedDirs := []string{trustedDir}
	pluginPath := filepath.Join(trustedDir, "plugin.so")

	if !IsPathInTrustedDirectory(pluginPath, trustedDirs) {
		t.Error("Path should be in trusted directory")
	}

	// Path not in trusted directory
	untrustedPath := filepath.Join(untrustedDir, "evil.so")
	if IsPathInTrustedDirectory(untrustedPath, trustedDirs) {
		t.Error("Path should not be in trusted directory")
	}
}

func TestIsPathInWorkingDirectory(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get cwd: %v", err)
	}

	// Path within cwd
	insidePath := filepath.Join(cwd, "plugins")
	inWorkDir, err := IsPathInWorkingDirectory(insidePath)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !inWorkDir {
		t.Error("Path inside cwd should return true")
	}

	// Path outside cwd
	outsidePath := "/etc"
	inWorkDir, err = IsPathInWorkingDirectory(outsidePath)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if inWorkDir {
		t.Error("Path outside cwd should return false")
	}
}

func TestPluginSecurityLogger(t *testing.T) {
	logger := NewPluginSecurityLogger()

	if len(logger.GetEvents()) != 0 {
		t.Error("New logger should have no events")
	}

	event := PluginLoadEvent{
		PluginPath: "/path/to/plugin.so",
		PluginName: "test-plugin",
		Checksum:   "a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3",
		Success:    true,
	}

	logger.LogLoadAttempt(event)

	events := logger.GetEvents()
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	if events[0].PluginName != "test-plugin" {
		t.Errorf("Expected plugin name 'test-plugin', got '%s'", events[0].PluginName)
	}

	if events[0].Timestamp.IsZero() {
		t.Error("Event timestamp should be set")
	}
}

func TestTruncateChecksum(t *testing.T) {
	// Short checksum
	short := truncateChecksum("abc123")
	if short != "abc123" {
		t.Errorf("Short checksum should not be truncated: %s", short)
	}

	// Long checksum
	long := truncateChecksum("a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3")
	if long != "a665a4592042..." {
		t.Errorf("Long checksum should be truncated: %s", long)
	}
}

func TestVerifyPlugin(t *testing.T) {
	tmpDir := t.TempDir()
	testPlugin := filepath.Join(tmpDir, "test.so")
	err := os.WriteFile(testPlugin, []byte("fake plugin content"), 0600)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test with default config (no verification required)
	config := DefaultSecurityConfig()
	allowlist := NewPluginAllowlist()

	event, err := VerifyPlugin(testPlugin, config, allowlist)
	if err != nil {
		t.Errorf("Should not error with default config: %v", err)
	}
	if event == nil {
		t.Fatal("Event should not be nil")
	}
	if !event.Success {
		t.Error("Event should be marked as success")
	}
	if event.SecurityWarning == "" {
		t.Error("Should have security warning about unverified plugin")
	}

	// Test with verification required but no allowlist
	config.RequireVerification = true
	config.AllowUnsignedPlugins = false

	_, err = VerifyPlugin(testPlugin, config, allowlist)
	if err == nil {
		t.Error("Should error when verification required but no allowlist")
	}
	if _, ok := err.(*PluginSecurityError); !ok {
		t.Error("Error should be PluginSecurityError")
	}

	// Test with allowlist containing correct checksum
	config.RequireVerification = false
	checksum, _ := CalculateFileChecksum(testPlugin)
	allowlist.entries["test.so"] = AllowlistEntry{
		Name:     "test.so",
		Checksum: checksum,
		Enabled:  true,
	}

	event, err = VerifyPlugin(testPlugin, config, allowlist)
	if err != nil {
		t.Errorf("Should not error with correct checksum: %v", err)
	}
	if !event.Success {
		t.Error("Event should be success")
	}

	// Test with wrong checksum in allowlist
	allowlist.entries["test.so"] = AllowlistEntry{
		Name:     "test.so",
		Checksum: "wrongchecksum1234567890123456789012345678901234567890123456789012",
		Enabled:  true,
	}

	_, err = VerifyPlugin(testPlugin, config, allowlist)
	if err == nil {
		t.Error("Should error with wrong checksum")
	}
}

func TestManagerWithSecurity(t *testing.T) {
	tmpDir := t.TempDir()

	// Test NewManagerWithSecurity with nil config
	manager, err := NewManagerWithSecurity(tmpDir, true, nil)
	if err != nil {
		t.Errorf("Should not error with nil config: %v", err)
	}
	if manager == nil {
		t.Fatal("Manager should not be nil")
	}
	if manager.securityConfig == nil {
		t.Error("Security config should be set to default")
	}

	// Test with valid allowlist file
	allowlistPath := filepath.Join(tmpDir, "allowlist.txt")
	content := "test.so:a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3\n"
	err = os.WriteFile(allowlistPath, []byte(content), 0600)
	if err != nil {
		t.Fatalf("Failed to write allowlist: %v", err)
	}

	config := &SecurityConfig{
		AllowlistPath: allowlistPath,
	}

	manager, err = NewManagerWithSecurity(tmpDir, true, config)
	if err != nil {
		t.Errorf("Should not error with valid allowlist: %v", err)
	}
	if manager.allowlist == nil || manager.allowlist.IsEmpty() {
		t.Error("Allowlist should be loaded")
	}
}

func TestManagerSetSecurityConfig(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(tmpDir, true, nil)

	// Test setting nil config
	err := manager.SetSecurityConfig(nil)
	if err != nil {
		t.Errorf("Should not error with nil config: %v", err)
	}
	if manager.securityConfig == nil {
		t.Error("Security config should be set to default")
	}

	// Test setting config with invalid allowlist path
	config := &SecurityConfig{
		AllowlistPath: filepath.Join(tmpDir, "invalid_allowlist.txt"),
	}

	// Create the file with invalid content
	err = os.WriteFile(config.AllowlistPath, []byte("invalid"), 0600)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	err = manager.SetSecurityConfig(config)
	if err == nil {
		t.Error("Should error with invalid allowlist")
	}
}

func TestManagerGetSecurityEvents(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(tmpDir, true, nil)

	events := manager.GetSecurityEvents()
	if events == nil {
		t.Error("Should return empty slice, not nil")
	}
	if len(events) != 0 {
		t.Error("Should have no events initially")
	}
}
