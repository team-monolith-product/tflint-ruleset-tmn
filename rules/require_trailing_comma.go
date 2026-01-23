// Rule: require_trailing_comma
//
// 리스트와 객체의 마지막 항목에 trailing comma를 강제합니다.
// 여러 줄에 걸쳐 작성된 리스트/객체에서만 검사합니다.
//
// @example GOOD
// locals {
//   items = [
//     "a",
//     "b",
//     "c",
//   ]
//   map = {
//     a = 1
//     b = 2
//   }
// }
//
// @example BAD
// locals {
//   items = [
//     "a",
//     "b",
//     "c"
//   ]
// }

package rules

import (
	"bytes"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

type RequireTrailingCommaRule struct {
	tflint.DefaultRule
}

func NewRequireTrailingCommaRule() *RequireTrailingCommaRule {
	return &RequireTrailingCommaRule{}
}

func (r *RequireTrailingCommaRule) Name() string {
	return "require_trailing_comma"
}

func (r *RequireTrailingCommaRule) Enabled() bool {
	return true
}

func (r *RequireTrailingCommaRule) Severity() tflint.Severity {
	return tflint.WARNING
}

func (r *RequireTrailingCommaRule) Link() string {
	return ""
}

func (r *RequireTrailingCommaRule) Check(runner tflint.Runner) error {
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

func (r *RequireTrailingCommaRule) checkFile(runner tflint.Runner, file *hcl.File) error {
	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return nil
	}

	return r.walkBody(runner, file.Bytes, body)
}

func (r *RequireTrailingCommaRule) walkBody(runner tflint.Runner, src []byte, body *hclsyntax.Body) error {
	for _, attr := range body.Attributes {
		if err := r.walkExpr(runner, src, attr.Expr); err != nil {
			return err
		}
	}

	for _, block := range body.Blocks {
		if err := r.walkBody(runner, src, block.Body); err != nil {
			return err
		}
	}

	return nil
}

func (r *RequireTrailingCommaRule) walkExpr(runner tflint.Runner, src []byte, expr hclsyntax.Expression) error {
	switch e := expr.(type) {
	case *hclsyntax.TupleConsExpr:
		if err := r.checkTuple(runner, src, e); err != nil {
			return err
		}
		for _, item := range e.Exprs {
			if err := r.walkExpr(runner, src, item); err != nil {
				return err
			}
		}

	case *hclsyntax.ObjectConsExpr:
		if err := r.checkObject(runner, src, e); err != nil {
			return err
		}
		for _, item := range e.Items {
			if err := r.walkExpr(runner, src, item.KeyExpr); err != nil {
				return err
			}
			if err := r.walkExpr(runner, src, item.ValueExpr); err != nil {
				return err
			}
		}

	case *hclsyntax.FunctionCallExpr:
		for _, arg := range e.Args {
			if err := r.walkExpr(runner, src, arg); err != nil {
				return err
			}
		}

	case *hclsyntax.ConditionalExpr:
		if err := r.walkExpr(runner, src, e.Condition); err != nil {
			return err
		}
		if err := r.walkExpr(runner, src, e.TrueResult); err != nil {
			return err
		}
		if err := r.walkExpr(runner, src, e.FalseResult); err != nil {
			return err
		}

	case *hclsyntax.ForExpr:
		if err := r.walkExpr(runner, src, e.CollExpr); err != nil {
			return err
		}
		if e.KeyExpr != nil {
			if err := r.walkExpr(runner, src, e.KeyExpr); err != nil {
				return err
			}
		}
		if err := r.walkExpr(runner, src, e.ValExpr); err != nil {
			return err
		}
		if e.CondExpr != nil {
			if err := r.walkExpr(runner, src, e.CondExpr); err != nil {
				return err
			}
		}

	case *hclsyntax.IndexExpr:
		if err := r.walkExpr(runner, src, e.Collection); err != nil {
			return err
		}
		if err := r.walkExpr(runner, src, e.Key); err != nil {
			return err
		}

	case *hclsyntax.SplatExpr:
		if err := r.walkExpr(runner, src, e.Source); err != nil {
			return err
		}
		if err := r.walkExpr(runner, src, e.Each); err != nil {
			return err
		}

	case *hclsyntax.RelativeTraversalExpr:
		if err := r.walkExpr(runner, src, e.Source); err != nil {
			return err
		}

	case *hclsyntax.ParenthesesExpr:
		if err := r.walkExpr(runner, src, e.Expression); err != nil {
			return err
		}

	case *hclsyntax.UnaryOpExpr:
		if err := r.walkExpr(runner, src, e.Val); err != nil {
			return err
		}

	case *hclsyntax.BinaryOpExpr:
		if err := r.walkExpr(runner, src, e.LHS); err != nil {
			return err
		}
		if err := r.walkExpr(runner, src, e.RHS); err != nil {
			return err
		}

	case *hclsyntax.TemplateExpr:
		for _, part := range e.Parts {
			if err := r.walkExpr(runner, src, part); err != nil {
				return err
			}
		}

	case *hclsyntax.TemplateWrapExpr:
		if err := r.walkExpr(runner, src, e.Wrapped); err != nil {
			return err
		}
	}

	return nil
}

func (r *RequireTrailingCommaRule) checkTuple(runner tflint.Runner, src []byte, tuple *hclsyntax.TupleConsExpr) error {
	if len(tuple.Exprs) == 0 {
		return nil
	}

	// Check if this is a multi-line tuple
	tupleRange := tuple.SrcRange
	if tupleRange.Start.Line == tupleRange.End.Line {
		return nil
	}

	lastExpr := tuple.Exprs[len(tuple.Exprs)-1]
	lastExprEnd := lastExpr.Range().End

	// Find the content between last expression end and closing bracket
	if !r.hasTrailingComma(src, lastExprEnd, tupleRange.End) {
		return runner.EmitIssueWithFix(
			r,
			"List should have a trailing comma after the last element",
			lastExpr.Range(),
			func(f tflint.Fixer) error {
				return f.InsertTextAfter(lastExpr.Range(), ",")
			},
		)
	}

	return nil
}

func (r *RequireTrailingCommaRule) checkObject(runner tflint.Runner, src []byte, obj *hclsyntax.ObjectConsExpr) error {
	if len(obj.Items) == 0 {
		return nil
	}

	// Check if this is a multi-line object
	objRange := obj.SrcRange
	if objRange.Start.Line == objRange.End.Line {
		return nil
	}

	lastItem := obj.Items[len(obj.Items)-1]
	lastItemEnd := lastItem.ValueExpr.Range().End

	// For objects in HCL, newline acts as separator, so trailing comma is optional
	// But we want to enforce it for consistency
	// Check if last item is on the same line as closing brace - if so, skip
	if lastItemEnd.Line == objRange.End.Line {
		return nil
	}

	// Check if there's a comma after the last item
	if !r.hasTrailingComma(src, lastItemEnd, objRange.End) {
		return runner.EmitIssueWithFix(
			r,
			"Object should have a trailing comma after the last element",
			lastItem.ValueExpr.Range(),
			func(f tflint.Fixer) error {
				return f.InsertTextAfter(lastItem.ValueExpr.Range(), ",")
			},
		)
	}

	return nil
}

func (r *RequireTrailingCommaRule) hasTrailingComma(src []byte, afterPos, beforePos hcl.Pos) bool {
	// Get the content between the positions
	lines := bytes.Split(src, []byte("\n"))

	// Look at the rest of the line after the expression
	if afterPos.Line > 0 && afterPos.Line <= len(lines) {
		line := lines[afterPos.Line-1]
		if afterPos.Column-1 < len(line) {
			restOfLine := line[afterPos.Column-1:]
			// Check if there's a comma in the rest of this line before any comment
			for _, ch := range restOfLine {
				if ch == ',' {
					return true
				}
				if ch == '#' || ch == '/' {
					break
				}
				if ch == ']' || ch == '}' {
					return false
				}
			}
		}
	}

	return false
}
