package rules

import (
	"bytes"
	"regexp"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// SortProviderAppMapRule checks that keys in provider_to_app_map nested objects
// are sorted alphabetically and have no extra blank lines between entries.
type SortProviderAppMapRule struct {
	tflint.DefaultRule
}

func NewSortProviderAppMapRule() *SortProviderAppMapRule {
	return &SortProviderAppMapRule{}
}

func (r *SortProviderAppMapRule) Name() string {
	return "sort_provider_app_map"
}

func (r *SortProviderAppMapRule) Enabled() bool {
	return true
}

func (r *SortProviderAppMapRule) Severity() tflint.Severity {
	return tflint.WARNING
}

func (r *SortProviderAppMapRule) Link() string {
	return ""
}

func (r *SortProviderAppMapRule) Check(runner tflint.Runner) error {
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := r.checkFile(runner, file); err != nil {
			return err
		}
	}

	return nil
}

func (r *SortProviderAppMapRule) checkFile(runner tflint.Runner, file *hcl.File) error {
	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return nil
	}

	// Check top-level attributes
	for name, attr := range body.Attributes {
		if name == "provider_to_app_map" {
			if err := r.checkProviderToAppMap(runner, file.Bytes, attr); err != nil {
				return err
			}
		}
	}

	// Check inside blocks (e.g., locals)
	for _, block := range body.Blocks {
		for name, attr := range block.Body.Attributes {
			if name == "provider_to_app_map" {
				if err := r.checkProviderToAppMap(runner, file.Bytes, attr); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (r *SortProviderAppMapRule) checkProviderToAppMap(runner tflint.Runner, src []byte, attr *hclsyntax.Attribute) error {
	objExpr, ok := attr.Expr.(*hclsyntax.ObjectConsExpr)
	if !ok {
		return nil
	}

	// Check each provider's nested object
	for _, item := range objExpr.Items {
		nestedObj, ok := item.ValueExpr.(*hclsyntax.ObjectConsExpr)
		if !ok {
			continue
		}

		if issue := r.checkNestedObjectSorting(nestedObj, src); issue != "" {
			// Get the provider key name for the message
			keyRange := item.KeyExpr.Range()
			keyName := string(keyRange.SliceBytes(src))

			// Capture variables for closure
			capturedObj := nestedObj

			err := runner.EmitIssueWithFix(
				r,
				"Keys in "+keyName+" should be sorted alphabetically with no extra blank lines",
				nestedObj.Range(),
				func(f tflint.Fixer) error {
					fixed := r.fixNestedObject(capturedObj, src)
					return f.ReplaceText(capturedObj.Range(), fixed)
				},
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *SortProviderAppMapRule) checkNestedObjectSorting(obj *hclsyntax.ObjectConsExpr, src []byte) string {
	if len(obj.Items) <= 1 {
		return ""
	}

	// Extract keys and check sorting
	var keys []string
	for _, item := range obj.Items {
		keyRange := item.KeyExpr.Range()
		key := string(keyRange.SliceBytes(src))
		keys = append(keys, key)
	}

	// Check if sorted
	sortedKeys := make([]string, len(keys))
	copy(sortedKeys, keys)
	sort.Strings(sortedKeys)

	for i, key := range keys {
		if key != sortedKeys[i] {
			return "not sorted"
		}
	}

	// Check for extra blank lines between items
	for i := 0; i < len(obj.Items)-1; i++ {
		currentEnd := obj.Items[i].ValueExpr.Range().End.Line
		nextStart := obj.Items[i+1].KeyExpr.Range().Start.Line
		if nextStart-currentEnd > 1 {
			return "extra blank lines"
		}
	}

	return ""
}

func (r *SortProviderAppMapRule) fixNestedObject(obj *hclsyntax.ObjectConsExpr, src []byte) string {
	type entry struct {
		key     string
		content string // includes preceding comments
	}

	lines := bytes.Split(src, []byte("\n"))
	var entries []entry

	// Extract each key-value pair with preceding comments
	for i, item := range obj.Items {
		keyRange := item.KeyExpr.Range()
		key := string(keyRange.SliceBytes(src))

		// Find the start line for this entry (including comments)
		var startLine int
		if i == 0 {
			// First item: start on the line after the opening brace
			startLine = obj.OpenRange.End.Line + 1
		} else {
			// Subsequent items: start on the line after previous item's value
			startLine = obj.Items[i-1].ValueExpr.Range().End.Line + 1
		}

		// Collect lines from startLine to end of this item's value
		endLine := item.ValueExpr.Range().End.Line

		var contentLines []string
		for lineNum := startLine; lineNum <= endLine; lineNum++ {
			if lineNum > 0 && lineNum <= len(lines) {
				line := string(lines[lineNum-1])
				// Skip empty lines before the comment/key, but keep comments
				trimmed := strings.TrimSpace(line)
				if trimmed == "" && lineNum < item.KeyExpr.Range().Start.Line {
					continue
				}
				contentLines = append(contentLines, line)
			}
		}

		entries = append(entries, entry{
			key:     key,
			content: strings.Join(contentLines, "\n"),
		})
	}

	// Sort by key
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].key < entries[j].key
	})

	// Build the fixed object
	var buf bytes.Buffer
	buf.WriteString("{\n")
	for i, e := range entries {
		buf.WriteString(e.content)
		if i < len(entries)-1 {
			buf.WriteString("\n")
		}
	}
	buf.WriteString("\n")
	// Close brace with parent indent (detect from opening brace line)
	openLine := obj.OpenRange.Start.Line
	if openLine > 0 && openLine <= len(lines) {
		lineContent := string(lines[openLine-1])
		parentIndent := ""
		for _, ch := range lineContent {
			if ch == ' ' || ch == '\t' {
				parentIndent += string(ch)
			} else {
				break
			}
		}
		buf.WriteString(parentIndent)
	}
	buf.WriteString("}")

	return buf.String()
}

func (r *SortProviderAppMapRule) detectIndent(obj *hclsyntax.ObjectConsExpr, src []byte) string {
	if len(obj.Items) == 0 {
		return "        " // default 8 spaces
	}

	// Get the line containing the first key
	firstKeyLine := obj.Items[0].KeyExpr.Range().Start.Line
	lines := bytes.Split(src, []byte("\n"))
	if firstKeyLine > 0 && firstKeyLine <= len(lines) {
		line := string(lines[firstKeyLine-1])
		// Extract leading whitespace
		re := regexp.MustCompile(`^(\s*)`)
		matches := re.FindStringSubmatch(line)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	return "        "
}
