package sitepages

import (
	"html/template"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// TestSplit tests the split function
func TestSplit(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		sep      string
		expected []string
	}{
		{"empty string", "", ",", []string{""}}, // strings.Split behavior
		{"no separator", "abc", ",", []string{"abc"}},
		{"simple split", "a,b,c", ",", []string{"a", "b", "c"}},
		{"trailing separator", "a,b,c,", ",", []string{"a", "b", "c", ""}}, // strings.Split behavior
		{"leading separator", ",a,b,c", ",", []string{"", "a", "b", "c"}}, // strings.Split behavior
		{"multiple separators", "a,,b,c", ",", []string{"a", "", "b", "c"}},
		{"different separator", "a|b|c", "|", []string{"a", "b", "c"}},
		{"empty string with separator", "", "|", []string{""}},
		{"separator only", ",", ",", []string{"", ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := split(tt.s, tt.sep)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("split(%q, %q) = %v, want %v", tt.s, tt.sep, result, tt.expected)
			}
		})
	}
}

// TestLoadAllTemplatePages tests the LoadAllTemplatePages function
func TestLoadAllTemplatePages(t *testing.T) {
	// Helper for creating temporary directories and files
	createTempFiles := func(t *testing.T, frontDirName, templateDirName string, frontFiles, templateFiles map[string]string) (string, string) {
		baseDir := t.TempDir()
		frontDir := filepath.Join(baseDir, frontDirName)
		templateDir := filepath.Join(baseDir, templateDirName)

		if err := os.Mkdir(frontDir, 0755); err != nil {
			t.Fatalf("Failed to create temp front dir: %v", err)
		}
		if err := os.Mkdir(templateDir, 0755); err != nil {
			t.Fatalf("Failed to create temp template dir: %v", err)
		}

		for name, content := range frontFiles {
			if err := os.WriteFile(filepath.Join(frontDir, name), []byte(content), 0644); err != nil {
				t.Fatalf("Failed to write temp front file %s: %v", name, err)
			}
		}
		for name, content := range templateFiles {
			if err := os.WriteFile(filepath.Join(templateDir, name), []byte(content), 0644); err != nil {
				t.Fatalf("Failed to write temp template file %s: %v", name, err)
			}
		}
		return frontDir, templateDir
	}

	// Default SetupTemplate function for tests
	defaultSetupFn := func(templ *template.Template) *template.Template {
		return templ
	}

	t.Run("successful_load", func(t *testing.T) {
		frontFiles := map[string]string{
			"page1.html": `{{define "content"}}Page 1 Content{{end}}`,
			"page2.html": `{{define "content"}}Page 2 Content{{end}}`,
		}
		templateFiles := map[string]string{
			"layout.html": `{{define "layout"}}Layout start {{template "content" .}} Layout end{{end}}`,
			"base.html":   `{{define "base"}}{{template "layout" .}}{{end}}`, // Assume this is the main base used by ParseGlob
		}
		frontDir, templateDir := createTempFiles(t, "front", "tmpl", frontFiles, templateFiles)

		// Note: LoadAllTemplatePages uses log.Fatal. If an error occurs, the test will stop.
		// We can only reliably test the success path here.
		templateMap := LoadAllTemplatePages(frontDir, templateDir, defaultSetupFn)

		if len(templateMap) != 2 {
			t.Fatalf("Expected 2 templates to be loaded, got %d", len(templateMap))
		}

		// Check page1.html
		tmpl1 := templateMap["page1.html"]
		if tmpl1 == nil {
			t.Fatal("page1.html not found in loaded templates")
		}
		// Check if it has the content from page1.html and the layout.html, via base.html
		// This means "base" should be a defined template in tmpl1.
		if tmpl1.Lookup("base") == nil {
			t.Error("page1.html does not seem to be correctly parsed with base.html")
		}
		if tmpl1.Lookup("layout") == nil {
			t.Error("page1.html does not seem to be correctly parsed with layout.html")
		}
		if tmpl1.Lookup("content") == nil { // "content" from page1.html itself
			t.Error("page1.html does not seem to have its own 'content' template defined")
		}
		
		// Try to execute to be more certain (optional, but good check)
		var buf strings.Builder
		err := tmpl1.ExecuteTemplate(&buf, "base", nil) // Execute the "base" which then calls "layout" then "content"
		if err != nil {
			t.Errorf("Error executing page1.html template: %v", err)
		}
		if !strings.Contains(buf.String(), "Page 1 Content") || !strings.Contains(buf.String(), "Layout start") {
			t.Errorf("page1.html execution output is not as expected: %s", buf.String())
		}


		// Check page2.html
		tmpl2 := templateMap["page2.html"]
		if tmpl2 == nil {
			t.Fatal("page2.html not found in loaded templates")
		}
		if tmpl2.Lookup("base") == nil {
			t.Error("page2.html does not seem to be correctly parsed with base.html")
		}
	})

	// Note: Testing error paths (ReadDir fails, ParseFiles/ParseGlob fails) is difficult
	// because LoadAllTemplatePages uses log.Fatal. These would terminate the test run.
	// To test these, the function would need to return an error.
	// For example, if os.ReadDir(frontFolder) fails, the test suite stops.
	// We can demonstrate this by trying to load from a non-existent directory,
	// but this test case itself would need to be run in a way that expects termination,
	// or it simply serves as documentation of current behavior.

	t.Run("front_dir_not_exists_causes_fatal", func(t *testing.T) {
		// This test documents that a non-existent frontFolder causes log.Fatal.
		// It's hard to assert this in a Go test without a separate process.
		// We are not actually running this part of the test in a way that it can pass CI if it fatals.
		// This is more of a note.
		// If you were to run: LoadAllTemplatePages("nonexistent_dir", "sometmpl", defaultSetupFn)
		// the test process would exit.
		t.Log("Skipping direct test for non-existent front_dir as it causes log.Fatal")
	})
	
	t.Run("template_parse_error_causes_fatal", func(t *testing.T) {
		// frontFiles := map[string]string{"page.html": `{{define "content"}}Valid page{{end}}`}
		// // Malformed template in templateFolder
		// templateFiles := map[string]string{"layout.html": `{{define "layout"}}Malformed {{template "content" .`} 
		// // frontDir, templateDir := createTempFiles(t, "front_parse_err", "tmpl_parse_err", frontFiles, templateFiles) // Unused

		// This would log.Fatal if called:
		// LoadAllTemplatePages(frontDir, templateDir, defaultSetupFn) 
		t.Log("Skipping direct test for template parse error as it causes log.Fatal")
	})
	
	t.Run("no_html_files_in_template_folder", func(t *testing.T) {
		frontFiles := map[string]string{"page.html": `{{define "content"}}Page Content{{end}}`}
		templateFiles := map[string]string{"not_html.txt": `some text`} // No .html files
		frontDir, templateDir := createTempFiles(t, "front_no_html", "tmpl_no_html", frontFiles, templateFiles)

		// template.ParseGlob(templateFolder + "*.html") will not find any files.
		// This is not an error for ParseGlob itself (it returns a nil error and an empty template if no files match).
		// The template 'tmpl' (parsed from frontFolder) will remain as is.
		templateMap := LoadAllTemplatePages(frontDir, templateDir, defaultSetupFn)
		if len(templateMap) != 1 {
			t.Fatalf("Expected 1 template, got %d", len(templateMap))
		}
		tmpl := templateMap["page.html"]
		if tmpl == nil {
			t.Fatal("page.html not found")
		}
		// It should only have the "content" from page.html, no layout stuff.
		if tmpl.Lookup("content") == nil {
			t.Error("page.html should have 'content' defined")
		}
		if tmpl.Lookup("layout") != nil || tmpl.Lookup("base") != nil {
			t.Error("page.html should not have layout templates if templateFolder had no .html files")
		}
		var buf strings.Builder
		// Executing "content" directly, as there's no layout to invoke it.
		// Or, if the page itself is expected to be a full template after ParseFiles.
		// The current LoadAllTemplatePages does `tmpl, err = tmpl.ParseGlob(...)`.
		// If ParseGlob is empty, `tmpl` is unchanged from `ParseFiles(frontFolder + filename.Name())`.
		// So, executing `tmpl.Name()` (which is "page.html") or a defined template like "content"
		// should work if "page.html" is self-contained or defines what it executes.
		// If "page.html" defines "content", we can execute that.
		err := tmpl.ExecuteTemplate(&buf, "content", nil)
		if err != nil {
			t.Errorf("Error executing template from page.html with no matching layout files: %v", err)
		}
		if !strings.Contains(buf.String(), "Page Content") {
			t.Errorf("Expected 'Page Content', got %s", buf.String())
		}
	})
}
