package tests

import (
	"strings"
	"testing"

	"github.com/team-monolith-product/tflint-ruleset-tmn/rules"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestSortAdditionalParametersDataRule(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name: "sorted by name should not trigger",
			content: `
provider_to_app_map = {
    aws = {
        user_rails = {
            additional_parameters = [
                {
                    name  = "react.env.REACT_APP_AI_URL"
                    value = "https://ai.example.com"
                },
                {
                    name  = "react.env.REACT_APP_LANDING_URL"
                    value = "https://landing.example.com"
                },
                {
                    name  = "react.ingress.hosts[0].host"
                    value = "example.com"
                },
            ]
        }
    }
}`,
			expected: 0,
		},
		{
			name: "unsorted by name should trigger",
			content: `
provider_to_app_map = {
    aws = {
        user_rails = {
            additional_parameters = [
                {
                    name  = "react.ingress.hosts[0].host"
                    value = "example.com"
                },
                {
                    name  = "react.env.REACT_APP_AI_URL"
                    value = "https://ai.example.com"
                },
                {
                    name  = "react.env.REACT_APP_LANDING_URL"
                    value = "https://landing.example.com"
                },
            ]
        }
    }
}`,
			expected: 1,
		},
		{
			name: "single element should not trigger",
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
			name: "inside locals block should work",
			content: `
locals {
    provider_to_app_map = {
        aws = {
            user_rails = {
                additional_parameters = [
                    {
                        name  = "zzz"
                        value = "val-z"
                    },
                    {
                        name  = "aaa"
                        value = "val-a"
                    },
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
			name: "multiple apps with unsorted parameters",
			content: `
provider_to_app_map = {
    aws = {
        user_rails = {
            additional_parameters = [
                {
                    name  = "zzz"
                    value = "val-z"
                },
                {
                    name  = "aaa"
                    value = "val-a"
                },
            ]
        }
        admin_rails = {
            additional_parameters = [
                {
                    name  = "yyy"
                    value = "val-y"
                },
                {
                    name  = "bbb"
                    value = "val-b"
                },
            ]
        }
    }
}`,
			expected: 2,
		},
	}

	rule := rules.NewSortAdditionalParametersDataRule()

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

func TestSortAdditionalParametersDataRule_Autofix(t *testing.T) {
	content := `
provider_to_app_map = {
    aws = {
        user_rails = {
            additional_parameters = [
                {
                    name  = "react.ingress.tls[0].hosts[0]"
                    value = local.dep.domain
                },
                {
                    name  = "react.ingress.hosts[0].host"
                    value = local.dep.domain
                },
                {
                    name  = "react.env.REACT_APP_JUPYTER_HUB_REPLICA"
                    value = local.jupyter_hub_replica_count
                },
                {
                    name  = "react.env.REACT_APP_AI_FASTAPI_URL"
                    value = "https://ai-fastapi.${local.dep.domain}"
                },
                {
                    name  = "react.env.REACT_APP_LANDING_URL"
                    value = "https://ai.${local.dep.domain}"
                },
            ]
        }
    }
}`

	rule := rules.NewSortAdditionalParametersDataRule()
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

	// Expected order by name:
	// react.env.REACT_APP_AI_FASTAPI_URL
	// react.env.REACT_APP_JUPYTER_HUB_REPLICA
	// react.env.REACT_APP_LANDING_URL
	// react.ingress.hosts[0].host
	// react.ingress.tls[0].hosts[0]
	aiIdx := indexOf(fixed, "REACT_APP_AI_FASTAPI_URL")
	jupyterIdx := indexOf(fixed, "REACT_APP_JUPYTER_HUB_REPLICA")
	landingIdx := indexOf(fixed, "REACT_APP_LANDING_URL")
	hostsIdx := indexOf(fixed, "react.ingress.hosts[0].host")
	tlsIdx := indexOf(fixed, "react.ingress.tls[0].hosts[0]")

	if aiIdx == -1 || jupyterIdx == -1 || landingIdx == -1 || hostsIdx == -1 || tlsIdx == -1 {
		t.Fatalf("missing expected names in fixed content:\n%s", fixed)
	}

	if !(aiIdx < jupyterIdx && jupyterIdx < landingIdx && landingIdx < hostsIdx && hostsIdx < tlsIdx) {
		t.Errorf("elements not properly sorted by name.\n"+
			"ai=%d, jupyter=%d, landing=%d, hosts=%d, tls=%d\nfixed:\n%s",
			aiIdx, jupyterIdx, landingIdx, hostsIdx, tlsIdx, fixed)
	}
}

func TestSortAdditionalParametersDataRule_Autofix_PreservesValues(t *testing.T) {
	content := `
provider_to_app_map = {
    aws = {
        user_rails = {
            additional_parameters = [
                {
                    name  = "zzz-param"
                    value = "zzz-value"
                },
                {
                    name  = "aaa-param"
                    value = "aaa-value"
                },
            ]
        }
    }
}`

	rule := rules.NewSortAdditionalParametersDataRule()
	runner := helper.TestRunner(t, map[string]string{
		"local_config.tf": content,
	})

	if err := rule.Check(runner); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	changes := runner.Changes()
	fixed := string(changes["local_config.tf"])

	// After fix, aaa-param should come before zzz-param
	aaaIdx := indexOf(fixed, "aaa-param")
	zzzIdx := indexOf(fixed, "zzz-param")

	if aaaIdx == -1 || zzzIdx == -1 {
		t.Fatalf("missing expected content in fixed output:\n%s", fixed)
	}

	if aaaIdx > zzzIdx {
		t.Errorf("aaa-param should come before zzz-param\nfixed:\n%s", fixed)
	}

	// Values should still be associated correctly
	aaaValueIdx := indexOf(fixed, "aaa-value")
	zzzValueIdx := indexOf(fixed, "zzz-value")

	if aaaValueIdx == -1 || zzzValueIdx == -1 {
		t.Fatalf("values were lost in fix:\n%s", fixed)
	}

	// aaa-value should appear after aaa-param and before zzz-param
	if !(aaaIdx < aaaValueIdx && aaaValueIdx < zzzIdx) {
		t.Errorf("aaa-value not properly associated with aaa-param\nfixed:\n%s", fixed)
	}
}

func TestSortAdditionalParametersDataRule_Autofix_LastElementNoTrailingComma(t *testing.T) {
	// Reproduces the bug where the last element has no trailing comma.
	// After sorting, it may become a middle element, causing "Missing item separator".
	content := `
provider_to_app_map = {
    aws = {
        user_rails = {
            additional_parameters = [
                {
                    name  = "ingress.tls[0].hosts[0]",
                    value = "class.example.com"
                },
                {
                    name  = "secretName"
                    value = "class-rails-master-key-secret"
                },
                {
                    name  = "env.AI_FASTAPI_URL"
                    value = "https://ai-fastapi.example.com"
                }
            ]
        }
    }
}`

	rule := rules.NewSortAdditionalParametersDataRule()
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

	// Expected order: env.AI_FASTAPI_URL < ingress.tls[0].hosts[0] < secretName
	aiIdx := indexOf(fixed, "env.AI_FASTAPI_URL")
	ingressIdx := indexOf(fixed, "ingress.tls[0].hosts[0]")
	secretIdx := indexOf(fixed, "secretName")

	if aiIdx == -1 || ingressIdx == -1 || secretIdx == -1 {
		t.Fatalf("missing expected names in fixed content:\n%s", fixed)
	}

	if !(aiIdx < ingressIdx && ingressIdx < secretIdx) {
		t.Errorf("elements not properly sorted by name.\n"+
			"ai=%d, ingress=%d, secret=%d\nfixed:\n%s",
			aiIdx, ingressIdx, secretIdx, fixed)
	}

	// Verify the fixed output is valid HCL by checking it doesn't have
	// consecutive closing/opening braces without commas
	if strings.Contains(fixed, "}\n                {") {
		t.Errorf("missing comma separator between elements in fixed output:\n%s", fixed)
	}
}
