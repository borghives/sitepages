package sitepages

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

// TestSaveSitePages tests the SaveSitePages function
func TestSaveSitePages(t *testing.T) {
	// Create a temporary file for saving
	tmpFile, err := os.CreateTemp("", "test_save_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name()) // Clean up
	tmpFilePath := tmpFile.Name()
	tmpFile.Close() // Close it so SaveSitePages can open and write to it.

	eventTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	pages := []SitePage{
		{
			LinkName: "testpage1",
			Title:    "Test Page 1",
			EventAt:  eventTime,
			// Populate other necessary fields if they affect serialization
		},
	}

	err = SaveSitePages(tmpFilePath, pages)
	if err != nil {
		t.Fatalf("SaveSitePages failed: %v", err)
	}

	// Read back the content and verify
	savedData, err := os.ReadFile(tmpFilePath)
	if err != nil {
		t.Fatalf("Failed to read back saved file: %v", err)
	}

	var loadedPages []SitePage
	err = json.Unmarshal(savedData, &loadedPages)
	if err != nil {
		t.Fatalf("Failed to unmarshal saved data: %v", err)
	}

	if len(loadedPages) != 1 {
		t.Fatalf("Expected 1 page to be loaded, got %d", len(loadedPages))
	}
	if loadedPages[0].LinkName != "testpage1" {
		t.Errorf("Expected LinkName 'testpage1', got '%s'", loadedPages[0].LinkName)
	}
	if !loadedPages[0].EventAt.Equal(eventTime) {
		// Time comparison can be tricky with JSON (timezones, precision).
		// Ensure EventAt is marshalled and unmarshalled consistently.
		// The default time.Time marshalling is RFC3339, which should be fine.
		t.Errorf("Expected EventAt %v, got %v", eventTime, loadedPages[0].EventAt)
	}
}

// TestLoadSitePages tests the LoadSitePages function
func TestLoadSitePages(t *testing.T) {
	eventTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	// Note: site_page.go's SitePage struct uses `EventAt` and `UpdatedTime`.
	// JSON tags for these are `event_at` and `updated_time` or `updated`.
	// The GenerateMomentString produces "2006-01-02 15:04" format, not RFC3339.
	// This will be an issue if EventAt is directly unmarshalled from a string produced by GenerateMomentString.
	// However, standard json marshalling of time.Time is RFC3339. Let's assume standard marshalling for EventAt.
	mockData := `[{"LinkName":"testpage1","Title":"Test Page 1","EventAt":"2024-01-01T10:00:00Z"}]`

	// Create a temporary file with mock data
	tmpFile, err := os.CreateTemp("", "test_load_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFilePath := tmpFile.Name()

	if _, err := tmpFile.Write([]byte(mockData)); err != nil {
		tmpFile.Close()
		t.Fatalf("Failed to write mock data to temp file: %v", err)
	}
	tmpFile.Close()

	loadedPages := LoadSitePages(tmpFilePath) // LoadSitePages calls log.Fatal on error.

	if len(loadedPages) != 1 {
		t.Fatalf("Expected 1 page, got %d", len(loadedPages))
	}
	page := loadedPages[0]

	if page.LinkName != "testpage1" {
		t.Errorf("Expected page LinkName 'testpage1', got '%s'", page.LinkName)
	}
	if !page.EventAt.Equal(eventTime) {
		t.Errorf("Expected EventAt %v, got %v", eventTime, page.EventAt)
	}

	// Test for non-existent file: LoadSitePages calls log.Fatal, so we can't directly test the error return.
	// To test this, we'd need to run it in a separate process or modify LoadSitePages to return an error.
	// For now, we skip directly testing non-existent file error path due to log.Fatal.
	// Same for corrupted JSON.
}


// TestGenerateMomentString tests the GenerateMomentString function
func TestGenerateMomentString(t *testing.T) {
	// The function under test uses time.Now(), so testing exact output is tricky.
	// We can test the format and that the time is recent.
	// Or, more robustly, mock time.Now() if possible, or test properties.
	// For this test, we'll check the format and that it's a valid time.
	// The function takes a coolDown duration.
	coolDown := 5 * time.Minute
	result := GenerateMomentString(coolDown)
	expectedFormat := "2006-01-02 15:04" // This is the format string used in the function

	parsedTime, err := time.Parse(expectedFormat, result)
	if err != nil {
		t.Fatalf("GenerateMomentString() produced a string that could not be parsed with its own format: %v. String was: %s", err, result)
	}

	// Check if the parsed time is roughly now + cooldown.
	// Allow for a small delta due to execution time.
	expectedTime := time.Now().UTC().Add(coolDown)
	delta := expectedTime.Sub(parsedTime)
	if delta < -1*time.Minute || delta > 1*time.Minute { // Allow 1 minute leeway
		t.Errorf("GenerateMomentString() result %v is too far from expected time %v (delta: %v)", parsedTime, expectedTime, delta)
	}
}

// TestParseMomentString tests the ParseMomementString function (note typo in original func name)
func TestParseMomementString(t *testing.T) {
	// The format used by ParseMomementString is "2006-01-02 15:04"
	timeStr := "2023-10-26 10:30"
	expected, _ := time.Parse("2006-01-02 15:04", timeStr) // Use same format for expected

	result, err := ParseMomementString(timeStr)
	if err != nil {
		t.Fatalf("ParseMomementString(%q) error: %v", timeStr, err)
	}
	if !result.Equal(expected) {
		t.Errorf("ParseMomementString(%q) = %v, want %v", timeStr, result, expected)
	}

	// Test invalid string
	invalidStr := "not-a-time-string"
	_, err = ParseMomementString(invalidStr)
	if err == nil {
		t.Errorf("ParseMomementString(%q) expected error, got nil", invalidStr)
	}
}

// TestCastRelationType tests the CastRelationType function
func TestCastRelationType(t *testing.T) {
	validTypes := map[string]RelationType{
		"bookmarked": RelationType_Bookmarked,
		"endorsed":   RelationType_Endorsed,
		"objected":   RelationType_Objected,
		"ignored":    RelationType_Ignored,
	}

	for s, expectedType := range validTypes {
		actualType := CastRelationType(s) // This function does not return an error
		if actualType != expectedType {
			t.Errorf("CastRelationType(%q) = %v, want %v", s, actualType, expectedType)
		}
	}

	// Test invalid type - should return RelationType_Generic
	invalidStr := "unknown_relation"
	expectedDefault := RelationType_Generic
	actualType := CastRelationType(invalidStr)
	if actualType != expectedDefault {
		t.Errorf("CastRelationType(%q) for invalid type = %v, want %v", invalidStr, actualType, expectedDefault)
	}
}

// TestCastRelationGraphType tests the CastRelationGraphType function
func TestCastRelationGraphType(t *testing.T) {
	validTypes := map[string]RelationGraphType{
		RelationGraphType_UserPage.String():    RelationGraphType_UserPage,
		RelationGraphType_UserComment.String(): RelationGraphType_UserComment,
		// "opaquerelation" is also a valid input string that should map to RelationGraphType_Opaque
		// but the function logic explicitly checks for UserPage and UserComment strings.
		// Let's test the defined cases and the default.
	}

	for s, expectedType := range validTypes {
		actualType := CastRelationGraphType(s) // This function does not return an error
		if actualType != expectedType {
			t.Errorf("CastRelationGraphType(%q) = %v, want %v", s, actualType, expectedType)
		}
	}

	// Test invalid type - should return RelationGraphType_Opaque
	invalidStr := "unknown_graph_type"
	expectedDefault := RelationGraphType_Opaque
	actualType := CastRelationGraphType(invalidStr)
	if actualType != expectedDefault {
		t.Errorf("CastRelationGraphType(%q) for invalid type = %v, want %v", invalidStr, actualType, expectedDefault)
	}
	
	// Test "opaquerelation" string explicitly if it's considered a valid input mapping to Opaque
	// Based on the switch, any string not matching UserPage or UserComment will result in Opaque.
	opaqueStr := "opaquerelation"
	actualOpaque := CastRelationGraphType(opaqueStr)
	if actualOpaque != RelationGraphType_Opaque {
		t.Errorf("CastRelationGraphType(%q) = %v, want %v", opaqueStr, actualOpaque, RelationGraphType_Opaque)
	}
}
