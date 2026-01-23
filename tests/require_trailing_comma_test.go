package tests

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/team-monolith-product/tflint-ruleset-tmn/rules"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestRequireTrailingCommaRule_List(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    helper.Issues
	}{
		{
			name: "multi-line list without trailing comma should trigger",
			content: `
locals {
  items = [
    "a",
    "b",
    "c"
  ]
}`,
			want: helper.Issues{
				{
					Rule:    rules.NewRequireTrailingCommaRule(),
					Message: "List should have a trailing comma after the last element",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 6, Column: 5, Byte: 44},
						End:      hcl.Pos{Line: 6, Column: 8, Byte: 47},
					},
				},
			},
		},
		{
			name: "multi-line list with trailing comma should not trigger",
			content: `
locals {
  items = [
    "a",
    "b",
    "c",
  ]
}`,
			want: helper.Issues{},
		},
		{
			name: "single-line list should not trigger",
			content: `
locals {
  items = ["a", "b", "c"]
}`,
			want: helper.Issues{},
		},
		{
			name: "empty list should not trigger",
			content: `
locals {
  items = []
}`,
			want: helper.Issues{},
		},
		{
			name: "single element list without trailing comma should trigger",
			content: `
locals {
  items = [
    "a"
  ]
}`,
			want: helper.Issues{
				{
					Rule:    rules.NewRequireTrailingCommaRule(),
					Message: "List should have a trailing comma after the last element",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 4, Column: 5, Byte: 24},
						End:      hcl.Pos{Line: 4, Column: 8, Byte: 27},
					},
				},
			},
		},
	}

	rule := rules.NewRequireTrailingCommaRule()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"resource.tf": tt.content,
			})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			helper.AssertIssues(t, tt.want, runner.Issues)
		})
	}
}

func TestRequireTrailingCommaRule_Object(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    helper.Issues
	}{
		{
			name: "multi-line object without trailing comma should trigger",
			content: `
locals {
  map = {
    a = 1
    b = 2
    c = 3
  }
}`,
			want: helper.Issues{
				{
					Rule:    rules.NewRequireTrailingCommaRule(),
					Message: "Object should have a trailing comma after the last element",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 6, Column: 9, Byte: 52},
						End:      hcl.Pos{Line: 6, Column: 10, Byte: 53},
					},
				},
			},
		},
		{
			name: "multi-line object with trailing comma should not trigger",
			content: `
locals {
  map = {
    a = 1
    b = 2
    c = 3,
  }
}`,
			want: helper.Issues{},
		},
		{
			name: "single-line object should not trigger",
			content: `
locals {
  map = { a = 1, b = 2 }
}`,
			want: helper.Issues{},
		},
	}

	rule := rules.NewRequireTrailingCommaRule()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"resource.tf": tt.content,
			})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			helper.AssertIssues(t, tt.want, runner.Issues)
		})
	}
}

func TestRequireTrailingCommaRule_Autofix_List(t *testing.T) {
	content := `
locals {
  items = [
    "a",
    "b",
    "c"
  ]
}`

	expected := `
locals {
  items = [
    "a",
    "b",
    "c",
  ]
}`

	rule := rules.NewRequireTrailingCommaRule()
	runner := helper.TestRunner(t, map[string]string{
		"resource.tf": content,
	})

	if err := rule.Check(runner); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	helper.AssertChanges(t, map[string]string{
		"resource.tf": expected,
	}, runner.Changes())
}

func TestRequireTrailingCommaRule_Autofix_Object(t *testing.T) {
	content := `
locals {
  map = {
    a = 1
    b = 2
  }
}`

	expected := `
locals {
  map = {
    a = 1
    b = 2,
  }
}`

	rule := rules.NewRequireTrailingCommaRule()
	runner := helper.TestRunner(t, map[string]string{
		"resource.tf": content,
	})

	if err := rule.Check(runner); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	helper.AssertChanges(t, map[string]string{
		"resource.tf": expected,
	}, runner.Changes())
}

func TestRequireTrailingCommaRule_NestedStructures(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    int // number of issues expected
	}{
		{
			name: "nested list in object without trailing comma",
			content: `
locals {
  config = {
    items = [
      "a",
      "b"
    ]
  }
}`,
			want: 2, // list and object both missing trailing comma
		},
		{
			name: "nested list in object with trailing commas",
			content: `
locals {
  config = {
    items = [
      "a",
      "b",
    ],
  }
}`,
			want: 0,
		},
	}

	rule := rules.NewRequireTrailingCommaRule()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"resource.tf": tt.content,
			})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if len(runner.Issues) != tt.want {
				t.Errorf("expected %d issues, got %d", tt.want, len(runner.Issues))
			}
		})
	}
}
