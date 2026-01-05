package rules

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// ForEachTosetRule checks for `for_each = { for x in list : x => x }` patterns
// and suggests using `toset()` instead.
type ForEachTosetRule struct {
	tflint.DefaultRule
}

func NewForEachTosetRule() *ForEachTosetRule {
	return &ForEachTosetRule{}
}

func (r *ForEachTosetRule) Name() string {
	return "foreach_toset"
}

func (r *ForEachTosetRule) Enabled() bool {
	return true
}

func (r *ForEachTosetRule) Severity() tflint.Severity {
	return tflint.WARNING
}

func (r *ForEachTosetRule) Link() string {
	return ""
}

func (r *ForEachTosetRule) Check(runner tflint.Runner) error {
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}

	for name, file := range files {
		if err := r.checkFile(runner, name, file); err != nil {
			return err
		}
	}

	return nil
}

func (r *ForEachTosetRule) checkFile(runner tflint.Runner, _ string, file *hcl.File) error {
	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return nil
	}

	for _, block := range body.Blocks {
		if err := r.checkBlock(runner, file.Bytes, block); err != nil {
			return err
		}
	}

	return nil
}

func (r *ForEachTosetRule) checkBlock(runner tflint.Runner, src []byte, block *hclsyntax.Block) error {
	// Check for_each attribute
	for name, attr := range block.Body.Attributes {
		if name == "for_each" {
			if err := r.checkForEachExpr(runner, src, attr); err != nil {
				return err
			}
		}
	}

	// Recursively check nested blocks
	for _, nestedBlock := range block.Body.Blocks {
		if err := r.checkBlock(runner, src, nestedBlock); err != nil {
			return err
		}
	}

	return nil
}

func (r *ForEachTosetRule) checkForEachExpr(runner tflint.Runner, src []byte, attr *hclsyntax.Attribute) error {
	forExpr, ok := attr.Expr.(*hclsyntax.ForExpr)
	if !ok {
		return nil
	}

	// Check if it's an object for expression (not a tuple)
	if forExpr.KeyExpr != nil {
		// Must be { for x in ... : x => x } pattern
		if r.isIdentityMapping(forExpr) {
			collRange := forExpr.CollExpr.Range()
			collBytes := collRange.SliceBytes(src)

			return runner.EmitIssueWithFix(
				r,
				"Use toset() instead of identity for expression",
				attr.Expr.Range(),
				func(f tflint.Fixer) error {
					return f.ReplaceText(attr.Expr.Range(), "toset(", string(collBytes), ")")
				},
			)
		}
	}

	return nil
}

// isIdentityMapping checks if a for expression is `{ for x in coll : x => x }`
func (r *ForEachTosetRule) isIdentityMapping(forExpr *hclsyntax.ForExpr) bool {
	// Must have a value variable (the iteration variable)
	if forExpr.ValVar == "" {
		return false
	}

	// Key variable must be empty (single variable iteration)
	if forExpr.KeyVar != "" {
		return false
	}

	// Check if key expression is just the iteration variable
	keyRef, keyOk := forExpr.KeyExpr.(*hclsyntax.ScopeTraversalExpr)
	if !keyOk || len(keyRef.Traversal) != 1 {
		return false
	}
	keyRoot, keyRootOk := keyRef.Traversal[0].(hcl.TraverseRoot)
	if !keyRootOk || keyRoot.Name != forExpr.ValVar {
		return false
	}

	// Check if value expression is just the iteration variable
	valRef, valOk := forExpr.ValExpr.(*hclsyntax.ScopeTraversalExpr)
	if !valOk || len(valRef.Traversal) != 1 {
		return false
	}
	valRoot, valRootOk := valRef.Traversal[0].(hcl.TraverseRoot)
	if !valRootOk || valRoot.Name != forExpr.ValVar {
		return false
	}

	// No condition in the for expression
	if forExpr.CondExpr != nil {
		return false
	}

	// Not grouping mode
	if forExpr.Group {
		return false
	}

	return true
}
