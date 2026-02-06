package tests

import (
	"strings"
	"testing"

	"github.com/team-monolith-product/tflint-ruleset-tmn/rules"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestRequireAdditionalParametersTrailingCommaRule(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name: "trailing comma present should not trigger",
			content: `
provider_to_app_map = {
    aws = {
        user_rails = {
            additional_parameters = [
                {
                    name  = "react.env.REACT_APP_AI_URL"
                    value = "https://ai.example.com"
                },
            ]
        }
    }
}`,
			expected: 0,
		},
		{
			name: "trailing comma missing should trigger",
			content: `
provider_to_app_map = {
    aws = {
        user_rails = {
            additional_parameters = [
                {
                    name  = "react.env.REACT_APP_AI_URL"
                    value = "https://ai.example.com"
                }
            ]
        }
    }
}`,
			expected: 1,
		},
		{
			name: "empty array should not trigger",
			content: `
provider_to_app_map = {
    aws = {
        user_rails = {
            additional_parameters = []
        }
    }
}`,
			expected: 0,
		},
		{
			name: "multiple elements with trailing comma should not trigger",
			content: `
provider_to_app_map = {
    aws = {
        user_rails = {
            additional_parameters = [
                {
                    name  = "aaa"
                    value = "val-a"
                },
                {
                    name  = "bbb"
                    value = "val-b"
                },
            ]
        }
    }
}`,
			expected: 0,
		},
		{
			name: "multiple elements without trailing comma on last should trigger",
			content: `
provider_to_app_map = {
    aws = {
        user_rails = {
            additional_parameters = [
                {
                    name  = "aaa"
                    value = "val-a"
                },
                {
                    name  = "bbb"
                    value = "val-b"
                }
            ]
        }
    }
}`,
			expected: 1,
		},
		{
			name: "inside locals block should work",
			content: `
locals {
    provider_to_app_map = {
        aws = {
            user_rails = {
                additional_parameters = [
                    {
                        name  = "aaa"
                        value = "val-a"
                    }
                ]
            }
        }
    }
}`,
			expected: 1,
		},
		{
			name: "no additional_parameters should not trigger",
			content: `
provider_to_app_map = {
    aws = {
        user_rails = {
            some_other_attr = "value"
        }
    }
}`,
			expected: 0,
		},
		{
			name: "multiple apps with missing trailing comma",
			content: `
provider_to_app_map = {
    aws = {
        user_rails = {
            additional_parameters = [
                {
                    name  = "aaa"
                    value = "val-a"
                }
            ]
        }
        admin_rails = {
            additional_parameters = [
                {
                    name  = "bbb"
                    value = "val-b"
                }
            ]
        }
    }
}`,
			expected: 2,
		},
		{
			name: "single element with trailing comma should not trigger",
			content: `
provider_to_app_map = {
    aws = {
        user_rails = {
            additional_parameters = [
                {
                    name  = "only-one"
                    value = "val"
                },
            ]
        }
    }
}`,
			expected: 0,
		},
	}

	rule := rules.NewRequireAdditionalParametersTrailingCommaRule()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"local_config.tf": tt.content,
			})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if len(runner.Issues) != tt.expected {
				t.Errorf("expected %d issues, got %d", tt.expected, len(runner.Issues))
				for _, issue := range runner.Issues {
					t.Logf("  issue: %s", issue.Message)
				}
			}
		})
	}
}

func TestRequireAdditionalParametersTrailingCommaRule_Autofix(t *testing.T) {
	content := `
provider_to_app_map = {
    aws = {
        user_rails = {
            additional_parameters = [
                {
                    name  = "react.env.REACT_APP_AI_URL"
                    value = "https://ai.example.com"
                }
            ]
        }
    }
}`

	rule := rules.NewRequireAdditionalParametersTrailingCommaRule()
	runner := helper.TestRunner(t, map[string]string{
		"local_config.tf": content,
	})

	if err := rule.Check(runner); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(runner.Issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(runner.Issues))
	}

	changes := runner.Changes()
	if len(changes) == 0 {
		t.Fatal("expected autofix changes")
	}

	fixedBytes := changes["local_config.tf"]
	if len(fixedBytes) == 0 {
		t.Fatal("expected changes to local_config.tf")
	}
	fixed := string(fixedBytes)

	// The closing brace of the last element should now have a trailing comma
	if !strings.Contains(fixed, "},") {
		t.Errorf("expected trailing comma in fixed content:\n%s", fixed)
	}

	// Verify the structure is preserved
	if !strings.Contains(fixed, `name  = "react.env.REACT_APP_AI_URL"`) {
		t.Errorf("name attribute lost in fixed content:\n%s", fixed)
	}
	if !strings.Contains(fixed, `value = "https://ai.example.com"`) {
		t.Errorf("value attribute lost in fixed content:\n%s", fixed)
	}
}

func TestRequireAdditionalParametersTrailingCommaRule_Autofix_MultipleElements(t *testing.T) {
	content := `
provider_to_app_map = {
    aws = {
        user_rails = {
            additional_parameters = [
                {
                    name  = "aaa"
                    value = "val-a"
                },
                {
                    name  = "bbb"
                    value = "val-b"
                }
            ]
        }
    }
}`

	rule := rules.NewRequireAdditionalParametersTrailingCommaRule()
	runner := helper.TestRunner(t, map[string]string{
		"local_config.tf": content,
	})

	if err := rule.Check(runner); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(runner.Issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(runner.Issues))
	}

	changes := runner.Changes()
	fixed := string(changes["local_config.tf"])

	// Both elements should now have trailing commas
	// The first already has one, the second should get one from the fix
	commaCount := strings.Count(fixed, "},")
	if commaCount != 2 {
		t.Errorf("expected 2 trailing commas in fixed content, got %d:\n%s", commaCount, fixed)
	}
}
