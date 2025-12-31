package tests

import (
	"testing"

	"github.com/team-monolith-product/tflint-ruleset-tmn/rules"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestProviderAppMapSortedRule(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int // number of issues
	}{
		{
			name: "sorted keys should not trigger",
			content: `
provider_to_app_map = {
    service = {
        aaa_config = {}
        bbb_config = {}
        ccc_config = {}
    }
}`,
			expected: 0,
		},
		{
			name: "unsorted keys should trigger",
			content: `
provider_to_app_map = {
    service = {
        ccc_config = {}
        aaa_config = {}
        bbb_config = {}
    }
}`,
			expected: 1,
		},
		{
			name: "extra blank lines should trigger",
			content: `
provider_to_app_map = {
    service = {
        aaa_config = {}

        bbb_config = {}
    }
}`,
			expected: 1,
		},
		{
			name: "single entry should not trigger",
			content: `
provider_to_app_map = {
    service = {
        only_config = {}
    }
}`,
			expected: 0,
		},
	}

	rule := rules.NewSortProviderAppMapRule()

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

func TestProviderAppMapSortedRule_Autofix(t *testing.T) {
	content := `
provider_to_app_map = {
    service = {
        zebra_config = {}
        alpha_config = {}

        beta_config = {}
    }
}`

	rule := rules.NewSortProviderAppMapRule()
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

	// Verify the fix contains sorted keys
	fixedBytes := changes["local_config.tf"]
	if len(fixedBytes) == 0 {
		t.Fatal("expected changes to local_config.tf")
	}
	fixed := string(fixedBytes)

	// Check that alpha comes before beta which comes before zebra
	alphaIdx := indexOf(fixed, "alpha_config")
	betaIdx := indexOf(fixed, "beta_config")
	zebraIdx := indexOf(fixed, "zebra_config")

	if alphaIdx == -1 || betaIdx == -1 || zebraIdx == -1 {
		t.Fatalf("missing expected keys in fixed content: %s", fixed)
	}

	if !(alphaIdx < betaIdx && betaIdx < zebraIdx) {
		t.Errorf("keys not properly sorted. alpha=%d, beta=%d, zebra=%d\nfixed:\n%s",
			alphaIdx, betaIdx, zebraIdx, fixed)
	}
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
