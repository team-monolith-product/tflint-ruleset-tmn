package tests

import (
	"testing"

	"github.com/team-monolith-product/tflint-ruleset-tmn/rules"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestSortAdditionalParametersDataRule(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int // number of issues
	}{
		{
			name: "sorted keys should not trigger",
			content: `
provider_to_app_map = {
    aws = {
        user_rails = {
            additional_parameters = [
                {
                    name = "param"
                    data = {
                        aaa-param = "value-a"
                        bbb-param = "value-b"
                        ccc-param = "value-c"
                    }
                }
            ]
        }
    }
}`,
			expected: 0,
		},
		{
			name: "unsorted keys should trigger",
			content: `
provider_to_app_map = {
    aws = {
        user_rails = {
            additional_parameters = [
                {
                    name = "param"
                    data = {
                        ccc-param = "value-c"
                        aaa-param = "value-a"
                        bbb-param = "value-b"
                    }
                }
            ]
        }
    }
}`,
			expected: 1,
		},
		{
			name: "single entry should not trigger",
			content: `
provider_to_app_map = {
    aws = {
        user_rails = {
            additional_parameters = [
                {
                    name = "param"
                    data = {
                        only-param = "value"
                    }
                }
            ]
        }
    }
}`,
			expected: 0,
		},
		{
			name: "multiple parameters with unsorted data should trigger multiple",
			content: `
provider_to_app_map = {
    aws = {
        user_rails = {
            additional_parameters = [
                {
                    name = "param1"
                    data = {
                        zzz-param = "value-z"
                        aaa-param = "value-a"
                    }
                },
                {
                    name = "param2"
                    data = {
                        yyy-param = "value-y"
                        bbb-param = "value-b"
                    }
                }
            ]
        }
    }
}`,
			expected: 2,
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
                        name = "param"
                        data = {
                            zzz-param = "value-z"
                            aaa-param = "value-a"
                        }
                    }
                ]
            }
        }
    }
}`,
			expected: 1,
		},
		{
			name: "empty data should not trigger",
			content: `
provider_to_app_map = {
    aws = {
        user_rails = {
            additional_parameters = [
                {
                    name = "param"
                    data = {}
                }
            ]
        }
    }
}`,
			expected: 0,
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
			name: "multiple apps with unsorted data",
			content: `
provider_to_app_map = {
    aws = {
        user_rails = {
            additional_parameters = [
                {
                    name = "param"
                    data = {
                        zzz = "value-z"
                        aaa = "value-a"
                    }
                }
            ]
        }
        admin_rails = {
            additional_parameters = [
                {
                    name = "param"
                    data = {
                        yyy = "value-y"
                        bbb = "value-b"
                    }
                }
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
                    name = "param"
                    data = {
                        zebra-param = "value-zebra"
                        alpha-param = "value-alpha"
                        beta-param  = "value-beta"
                    }
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
	if len(changes) == 0 {
		t.Fatal("expected autofix changes")
	}

	fixedBytes := changes["local_config.tf"]
	if len(fixedBytes) == 0 {
		t.Fatal("expected changes to local_config.tf")
	}
	fixed := string(fixedBytes)

	// Check that alpha comes before beta which comes before zebra
	alphaIdx := indexOf(fixed, "alpha-param")
	betaIdx := indexOf(fixed, "beta-param")
	zebraIdx := indexOf(fixed, "zebra-param")

	if alphaIdx == -1 || betaIdx == -1 || zebraIdx == -1 {
		t.Fatalf("missing expected keys in fixed content: %s", fixed)
	}

	if !(alphaIdx < betaIdx && betaIdx < zebraIdx) {
		t.Errorf("keys not properly sorted. alpha=%d, beta=%d, zebra=%d\nfixed:\n%s",
			alphaIdx, betaIdx, zebraIdx, fixed)
	}
}

func TestSortAdditionalParametersDataRule_RealWorldExample(t *testing.T) {
	content := `
locals {
    provider_to_app_map = {
        provider_a = {
            user_rails = {
                additional_secrets = []
                additional_parameters = [
                    {
                        name = "user-rails-params"
                        data = {
                            redis-url          = "redis://localhost:6379"
                            database-url       = "postgres://localhost:5432"
                            app-name           = "user-rails"
                            smtp-server        = "smtp.example.com"
                            log-level          = "info"
                        }
                    }
                ]
            }
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

	// Should trigger because keys are not sorted (app-name should come first, then database-url, etc.)
	if len(runner.Issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(runner.Issues))
	}

	changes := runner.Changes()
	fixed := string(changes["local_config.tf"])

	// Verify sorting: app-name < database-url < log-level < redis-url < smtp-server
	appNameIdx := indexOf(fixed, "app-name")
	databaseIdx := indexOf(fixed, "database-url")
	logLevelIdx := indexOf(fixed, "log-level")
	redisIdx := indexOf(fixed, "redis-url")
	smtpIdx := indexOf(fixed, "smtp-server")

	if !(appNameIdx < databaseIdx && databaseIdx < logLevelIdx && logLevelIdx < redisIdx && redisIdx < smtpIdx) {
		t.Errorf("keys not properly sorted in real-world example\nfixed:\n%s", fixed)
	}
}
