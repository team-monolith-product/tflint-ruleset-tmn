// Rule: sort_additional_secrets_data
//
// `additional_secrets[*].data` 내 키를 알파벳순으로 정렬합니다.
//
// @example GOOD
// locals {
//   provider_to_app_map = {
//     aws = {
//       user_rails = {
//         additional_secrets = [
//           {
//             name = "secret"
//             data = {
//               aaa-secret = local.secrets["aaa"]
//               bbb-secret = local.secrets["bbb"]
//               ccc-secret = local.secrets["ccc"]
//             }
//           }
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
//         additional_secrets = [
//           {
//             name = "secret"
//             data = {
//               ccc-secret = local.secrets["ccc"]
//               aaa-secret = local.secrets["aaa"]
//               bbb-secret = local.secrets["bbb"]
//             }
//           }
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

type SortAdditionalSecretsDataRule struct {
	tflint.DefaultRule
}

func NewSortAdditionalSecretsDataRule() *SortAdditionalSecretsDataRule {
	return &SortAdditionalSecretsDataRule{}
}

func (r *SortAdditionalSecretsDataRule) Name() string {
	return "sort_additional_secrets_data"
}

func (r *SortAdditionalSecretsDataRule) Enabled() bool {
	return true
}

func (r *SortAdditionalSecretsDataRule) Severity() tflint.Severity {
	return tflint.WARNING
}

func (r *SortAdditionalSecretsDataRule) Link() string {
	return ""
}

func (r *SortAdditionalSecretsDataRule) Check(runner tflint.Runner) error {
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

func (r *SortAdditionalSecretsDataRule) checkFile(runner tflint.Runner, file *hcl.File) error {
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

func (r *SortAdditionalSecretsDataRule) checkProviderToAppMap(runner tflint.Runner, src []byte, expr hclsyntax.Expression) error {
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

			// Find additional_secrets attribute
			for _, attrItem := range appObj.Items {
				keyRange := attrItem.KeyExpr.Range()
				key := string(keyRange.SliceBytes(src))
				// Remove quotes if present
				key = strings.Trim(key, "\"")

				if key == "additional_secrets" {
					if err := r.checkAdditionalSecrets(runner, src, attrItem.ValueExpr); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (r *SortAdditionalSecretsDataRule) checkAdditionalSecrets(runner tflint.Runner, src []byte, expr hclsyntax.Expression) error {
	tupleExpr, ok := expr.(*hclsyntax.TupleConsExpr)
	if !ok {
		return nil
	}

	// Iterate through each secret object in the array
	for _, elemExpr := range tupleExpr.Exprs {
		secretObj, ok := elemExpr.(*hclsyntax.ObjectConsExpr)
		if !ok {
			continue
		}

		// Find data attribute
		for _, attrItem := range secretObj.Items {
			keyRange := attrItem.KeyExpr.Range()
			key := string(keyRange.SliceBytes(src))
			key = strings.Trim(key, "\"")

			if key == "data" {
				dataObj, ok := attrItem.ValueExpr.(*hclsyntax.ObjectConsExpr)
				if !ok {
					continue
				}

				if err := r.checkDataObjectSorting(runner, src, dataObj); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (r *SortAdditionalSecretsDataRule) checkDataObjectSorting(runner tflint.Runner, src []byte, dataObj *hclsyntax.ObjectConsExpr) error {
	if len(dataObj.Items) <= 1 {
		return nil
	}

	// Extract keys
	var keys []string
	for _, item := range dataObj.Items {
		keyRange := item.KeyExpr.Range()
		key := string(keyRange.SliceBytes(src))
		keys = append(keys, key)
	}

	// Check if sorted
	sortedKeys := make([]string, len(keys))
	copy(sortedKeys, keys)
	sort.Strings(sortedKeys)

	needsFix := false
	for i, key := range keys {
		if key != sortedKeys[i] {
			needsFix = true
			break
		}
	}

	if !needsFix {
		return nil
	}

	// Capture for closure
	capturedObj := dataObj

	return runner.EmitIssueWithFix(
		r,
		"Keys in additional_secrets data should be sorted alphabetically",
		dataObj.Range(),
		func(f tflint.Fixer) error {
			fixed := r.fixDataObject(capturedObj, src)
			return f.ReplaceText(capturedObj.Range(), fixed)
		},
	)
}

func (r *SortAdditionalSecretsDataRule) fixDataObject(obj *hclsyntax.ObjectConsExpr, src []byte) string {
	if len(obj.Items) == 0 {
		return string(obj.Range().SliceBytes(src))
	}

	type entry struct {
		key     string
		content string
	}

	lines := bytes.Split(src, []byte("\n"))
	var entries []entry

	// Extract each key-value pair
	for i, item := range obj.Items {
		keyRange := item.KeyExpr.Range()
		key := string(keyRange.SliceBytes(src))

		var startLine int
		if i == 0 {
			startLine = obj.OpenRange.End.Line + 1
		} else {
			startLine = obj.Items[i-1].ValueExpr.Range().End.Line + 1
		}
		endLine := item.ValueExpr.Range().End.Line

		var contentLines []string
		for lineNum := startLine; lineNum <= endLine; lineNum++ {
			if lineNum > 0 && lineNum <= len(lines) {
				line := string(lines[lineNum-1])
				trimmed := strings.TrimSpace(line)
				if trimmed == "" {
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

	// Close brace with parent indent
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
