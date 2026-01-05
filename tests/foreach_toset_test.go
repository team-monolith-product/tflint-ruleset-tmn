package tests

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/team-monolith-product/tflint-ruleset-tmn/rules"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestForEachTosetRule(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    helper.Issues
	}{
		{
			name: "identity for expression should trigger",
			content: `
resource "aws_instance" "example" {
  for_each = { for x in var.instances : x => x }
  ami      = "ami-12345"
}`,
			want: helper.Issues{
				{
					Rule:    rules.NewForEachTosetRule(),
					Message: "Use toset() instead of identity for expression",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 14, Byte: 50},
						End:      hcl.Pos{Line: 3, Column: 49, Byte: 85},
					},
				},
			},
		},
		{
			name: "toset should not trigger",
			content: `
resource "aws_instance" "example" {
  for_each = toset(var.instances)
  ami      = "ami-12345"
}`,
			want: helper.Issues{},
		},
		{
			name: "non-identity for expression should not trigger",
			content: `
resource "aws_instance" "example" {
  for_each = { for x in var.instances : x.name => x }
  ami      = "ami-12345"
}`,
			want: helper.Issues{},
		},
		{
			name: "for expression with condition should not trigger",
			content: `
resource "aws_instance" "example" {
  for_each = { for x in var.instances : x => x if x != "" }
  ami      = "ami-12345"
}`,
			want: helper.Issues{},
		},
	}

	rule := rules.NewForEachTosetRule()

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

func TestForEachTosetRule_Autofix(t *testing.T) {
	content := `
resource "aws_instance" "example" {
  for_each = { for x in var.instances : x => x }
  ami      = "ami-12345"
}`

	expected := `
resource "aws_instance" "example" {
  for_each = toset(var.instances)
  ami      = "ami-12345"
}`

	rule := rules.NewForEachTosetRule()
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
