// Rule: sort_additional_parameters_data
//
// `additional_parameters` 배열 내 요소를 `name` 값 기준 알파벳순으로 정렬합니다.
//
// @example GOOD
// locals {
//   provider_to_app_map = {
//     aws = {
//       user_rails = {
//         additional_parameters = [
//           {
//             name  = "react.env.REACT_APP_AI_URL"
//             value = "https://ai.example.com"
//           },
//           {
//             name  = "react.env.REACT_APP_LANDING_URL"
//             value = "https://landing.example.com"
//           },
//           {
//             name  = "react.ingress.hosts[0].host"
//             value = "example.com"
//           },
//         ]
//       }
//     }
//   }
// }
//
// @example BAD
// locals {
//   provider_to_app_map = {
//     aws = {
//       user_rails = {
//         additional_parameters = [
//           {
//             name  = "react.ingress.hosts[0].host"
//             value = "example.com"
//           },
//           {
//             name  = "react.env.REACT_APP_AI_URL"
//             value = "https://ai.example.com"
//           },
//           {
//             name  = "react.env.REACT_APP_LANDING_URL"
//             value = "https://landing.example.com"
//           },
//         ]
//       }
//     }
//   }
// }

package rules

import (
	"bytes"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

type SortAdditionalParametersDataRule struct {
	tflint.DefaultRule
}

func NewSortAdditionalParametersDataRule() *SortAdditionalParametersDataRule {
	return &SortAdditionalParametersDataRule{}
}

func (r *SortAdditionalParametersDataRule) Name() string {
	return "sort_additional_parameters_data"
}

func (r *SortAdditionalParametersDataRule) Enabled() bool {
	return true
}

func (r *SortAdditionalParametersDataRule) Severity() tflint.Severity {
	return tflint.WARNING
}

func (r *SortAdditionalParametersDataRule) Link() string {
	return ""
}

func (r *SortAdditionalParametersDataRule) Check(runner tflint.Runner) error {
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

func (r *SortAdditionalParametersDataRule) checkFile(runner tflint.Runner, file *hcl.File) error {
	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return nil
	}

	// Check top-level attributes
	for name, attr := range body.Attributes {
		if name == "provider_to_app_map" {
			if err := r.checkProviderToAppMap(runner, file.Bytes, attr.Expr); err != nil {
				return err
			}
		}
	}

	// Check inside blocks (e.g., locals)
	for _, block := range body.Blocks {
		for name, attr := range block.Body.Attributes {
			if name == "provider_to_app_map" {
				if err := r.checkProviderToAppMap(runner, file.Bytes, attr.Expr); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (r *SortAdditionalParametersDataRule) checkProviderToAppMap(runner tflint.Runner, src []byte, expr hclsyntax.Expression) error {
	objExpr, ok := expr.(*hclsyntax.ObjectConsExpr)
	if !ok {
		return nil
	}

	// Iterate through providers (e.g., aws, gcp)
	for _, providerItem := range objExpr.Items {
		providerObj, ok := providerItem.ValueExpr.(*hclsyntax.ObjectConsExpr)
		if !ok {
			continue
		}

		// Iterate through apps (e.g., user_rails, admin_rails)
		for _, appItem := range providerObj.Items {
			appObj, ok := appItem.ValueExpr.(*hclsyntax.ObjectConsExpr)
			if !ok {
				continue
			}

			// Find additional_parameters attribute
			for _, attrItem := range appObj.Items {
				keyRange := attrItem.KeyExpr.Range()
				key := string(keyRange.SliceBytes(src))
				key = strings.Trim(key, "\"")

				if key == "additional_parameters" {
					if err := r.checkAdditionalParameters(runner, src, attrItem.ValueExpr); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (r *SortAdditionalParametersDataRule) checkAdditionalParameters(runner tflint.Runner, src []byte, expr hclsyntax.Expression) error {
	tupleExpr, ok := expr.(*hclsyntax.TupleConsExpr)
	if !ok {
		return nil
	}

	if len(tupleExpr.Exprs) <= 1 {
		return nil
	}

	// Extract name values from each element
	names := make([]string, 0, len(tupleExpr.Exprs))
	for _, elemExpr := range tupleExpr.Exprs {
		name := r.extractName(elemExpr, src)
		names = append(names, name)
	}

	// Check if sorted
	sorted := make([]string, len(names))
	copy(sorted, names)
	sort.Strings(sorted)

	needsFix := false
	for i, name := range names {
		if name != sorted[i] {
			needsFix = true
			break
		}
	}

	if !needsFix {
		return nil
	}

	capturedTuple := tupleExpr

	return runner.EmitIssueWithFix(
		r,
		"Elements in additional_parameters should be sorted by name",
		tupleExpr.Range(),
		func(f tflint.Fixer) error {
			fixed := r.fixTuple(capturedTuple, src)
			return f.ReplaceText(capturedTuple.Range(), fixed)
		},
	)
}

// extractName extracts the "name" attribute value from an object expression
func (r *SortAdditionalParametersDataRule) extractName(expr hclsyntax.Expression, src []byte) string {
	objExpr, ok := expr.(*hclsyntax.ObjectConsExpr)
	if !ok {
		return ""
	}

	for _, item := range objExpr.Items {
		keyRange := item.KeyExpr.Range()
		key := string(keyRange.SliceBytes(src))
		key = strings.Trim(key, "\"")

		if key == "name" {
			valRange := item.ValueExpr.Range()
			val := string(valRange.SliceBytes(src))
			return strings.Trim(val, "\"")
		}
	}

	return ""
}

func (r *SortAdditionalParametersDataRule) fixTuple(tuple *hclsyntax.TupleConsExpr, src []byte) string {
	if len(tuple.Exprs) == 0 {
		return string(tuple.Range().SliceBytes(src))
	}

	lines := bytes.Split(src, []byte("\n"))

	type entry struct {
		name    string
		content string
	}

	var entries []entry

	for i, elemExpr := range tuple.Exprs {
		name := r.extractName(elemExpr, src)

		// Determine the line range for this element
		elemRange := elemExpr.Range()
		startLine := elemRange.Start.Line
		endLine := elemRange.End.Line

		// Include trailing comma on the same line or next line
		if endLine > 0 && endLine <= len(lines) {
			lineAfterClose := string(lines[endLine-1])
			// Check if there's a comma after the closing brace on the same line
			endCol := elemRange.End.Column - 1
			if endCol < len(lineAfterClose) {
				rest := strings.TrimSpace(lineAfterClose[endCol:])
				if strings.HasPrefix(rest, ",") {
					// comma is on the same line, already included
				}
			}
		}

		// For elements after the first, check if there are blank lines before
		var contentLines []string
		actualStart := startLine
		if i > 0 {
			prevEnd := tuple.Exprs[i-1].Range().End.Line
			// Include lines between previous element and this one (blank lines)
			for lineNum := prevEnd + 1; lineNum < startLine; lineNum++ {
				if lineNum > 0 && lineNum <= len(lines) {
					trimmed := strings.TrimSpace(string(lines[lineNum-1]))
					if trimmed == "" || trimmed == "," {
						// skip inter-element whitespace/commas
						continue
					}
				}
			}
			_ = actualStart
		}

		// Extract the element content from source
		for lineNum := startLine; lineNum <= endLine; lineNum++ {
			if lineNum > 0 && lineNum <= len(lines) {
				contentLines = append(contentLines, string(lines[lineNum-1]))
			}
		}

		// Handle trailing comma: check if the last line has a comma after the closing brace
		if len(contentLines) > 0 {
			lastLine := contentLines[len(contentLines)-1]
			endCol := elemRange.End.Column - 1
			if endCol < len(lastLine) {
				rest := strings.TrimSpace(lastLine[endCol:])
				if !strings.HasPrefix(rest, ",") {
					// No trailing comma on this line, check if comma is on the next line
					nextLine := endLine + 1
					if nextLine > 0 && nextLine <= len(lines) {
						trimmed := strings.TrimSpace(string(lines[nextLine-1]))
						if trimmed == "," {
							contentLines = append(contentLines, string(lines[nextLine-1]))
						}
					}
				}
			}
		}

		// Normalize: strip trailing comma from captured content
		// so we can consistently add commas during reconstruction
		content := strings.Join(contentLines, "\n")
		content = strings.TrimRight(content, " \t\n")
		if strings.HasSuffix(content, ",") {
			content = content[:len(content)-1]
		}

		entries = append(entries, entry{
			name:    name,
			content: content,
		})
	}

	// Sort by name
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].name < entries[j].name
	})

	// Determine indentation from the opening bracket line
	openLine := tuple.OpenRange.Start.Line
	baseIndent := ""
	if openLine > 0 && openLine <= len(lines) {
		lineContent := string(lines[openLine-1])
		for _, ch := range lineContent {
			if ch == ' ' || ch == '\t' {
				baseIndent += string(ch)
			} else {
				break
			}
		}
	}

	// Build the fixed tuple
	var buf bytes.Buffer
	buf.WriteString("[\n")
	for _, e := range entries {
		buf.WriteString(e.content)
		buf.WriteString(",\n")
	}
	buf.WriteString(baseIndent)
	buf.WriteString("]")

	return buf.String()
}
