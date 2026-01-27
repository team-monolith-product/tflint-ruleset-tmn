// Rule: require_additional_parameters_trailing_comma
//
// `additional_parameters` 배열의 마지막 요소 뒤에 trailing comma를 요구합니다.
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
//             name  = "react.env.REACT_APP_AI_URL"
//             value = "https://ai.example.com"
//           }
//         ]
//       }
//     }
//   }
// }

package rules

import (
	"bytes"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

type RequireAdditionalParametersTrailingCommaRule struct {
	tflint.DefaultRule
}

func NewRequireAdditionalParametersTrailingCommaRule() *RequireAdditionalParametersTrailingCommaRule {
	return &RequireAdditionalParametersTrailingCommaRule{}
}

func (r *RequireAdditionalParametersTrailingCommaRule) Name() string {
	return "require_additional_parameters_trailing_comma"
}

func (r *RequireAdditionalParametersTrailingCommaRule) Enabled() bool {
	return true
}

func (r *RequireAdditionalParametersTrailingCommaRule) Severity() tflint.Severity {
	return tflint.WARNING
}

func (r *RequireAdditionalParametersTrailingCommaRule) Link() string {
	return ""
}

func (r *RequireAdditionalParametersTrailingCommaRule) Check(runner tflint.Runner) error {
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

func (r *RequireAdditionalParametersTrailingCommaRule) checkFile(runner tflint.Runner, file *hcl.File) error {
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

func (r *RequireAdditionalParametersTrailingCommaRule) checkProviderToAppMap(runner tflint.Runner, src []byte, expr hclsyntax.Expression) error {
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

func (r *RequireAdditionalParametersTrailingCommaRule) checkAdditionalParameters(runner tflint.Runner, src []byte, expr hclsyntax.Expression) error {
	tupleExpr, ok := expr.(*hclsyntax.TupleConsExpr)
	if !ok {
		return nil
	}

	if len(tupleExpr.Exprs) == 0 {
		return nil
	}

	lastElem := tupleExpr.Exprs[len(tupleExpr.Exprs)-1]
	lastElemEnd := lastElem.Range().End

	// Check if there's a trailing comma after the last element on the same line
	lines := bytes.Split(src, []byte("\n"))
	hasComma := false

	endLine := lastElemEnd.Line
	endCol := lastElemEnd.Column - 1 // Convert to 0-indexed

	if endLine > 0 && endLine <= len(lines) {
		lineContent := string(lines[endLine-1])
		if endCol < len(lineContent) {
			rest := strings.TrimSpace(lineContent[endCol:])
			if strings.HasPrefix(rest, ",") {
				hasComma = true
			}
		}
	}

	if hasComma {
		return nil
	}

	capturedLastElem := lastElem
	return runner.EmitIssueWithFix(
		r,
		"additional_parameters list should have a trailing comma after the last element",
		lastElem.Range(),
		func(f tflint.Fixer) error {
			elemContent := string(capturedLastElem.Range().SliceBytes(src))
			return f.ReplaceText(capturedLastElem.Range(), elemContent+",")
		},
	)
}
