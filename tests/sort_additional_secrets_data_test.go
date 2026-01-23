package tests

import (
	"testing"

	"github.com/team-monolith-product/tflint-ruleset-tmn/rules"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestSortAdditionalSecretsDataRule(t *testing.T) {
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
            additional_secrets = [
                {
                    name = "secret"
                    data = {
                        aaa-secret = local.secrets["aaa"]
                        bbb-secret = local.secrets["bbb"]
                        ccc-secret = local.secrets["ccc"]
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
            additional_secrets = [
                {
                    name = "secret"
                    data = {
                        ccc-secret = local.secrets["ccc"]
                        aaa-secret = local.secrets["aaa"]
                        bbb-secret = local.secrets["bbb"]
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
            additional_secrets = [
                {
                    name = "secret"
                    data = {
                        only-secret = local.secrets["only"]
                    }
                }
            ]
        }
    }
}`,
			expected: 0,
		},
		{
			name: "multiple secrets with unsorted data should trigger multiple",
			content: `
provider_to_app_map = {
    aws = {
        user_rails = {
            additional_secrets = [
                {
                    name = "secret1"
                    data = {
                        zzz-secret = local.secrets["zzz"]
                        aaa-secret = local.secrets["aaa"]
                    }
                },
                {
                    name = "secret2"
                    data = {
                        yyy-secret = local.secrets["yyy"]
                        bbb-secret = local.secrets["bbb"]
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
                additional_secrets = [
                    {
                        name = "secret"
                        data = {
                            zzz-secret = local.secrets["zzz"]
                            aaa-secret = local.secrets["aaa"]
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
            additional_secrets = [
                {
                    name = "secret"
                    data = {}
                }
            ]
        }
    }
}`,
			expected: 0,
		},
		{
			name: "no additional_secrets should not trigger",
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
            additional_secrets = [
                {
                    name = "secret"
                    data = {
                        zzz = local.secrets["zzz"]
                        aaa = local.secrets["aaa"]
                    }
                }
            ]
        }
        admin_rails = {
            additional_secrets = [
                {
                    name = "secret"
                    data = {
                        yyy = local.secrets["yyy"]
                        bbb = local.secrets["bbb"]
                    }
                }
            ]
        }
    }
}`,
			expected: 2,
		},
	}

	rule := rules.NewSortAdditionalSecretsDataRule()

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

func TestSortAdditionalSecretsDataRule_Autofix(t *testing.T) {
	content := `
provider_to_app_map = {
    aws = {
        user_rails = {
            additional_secrets = [
                {
                    name = "secret"
                    data = {
                        zebra-secret = local.secrets["zebra"]
                        alpha-secret = local.secrets["alpha"]
                        beta-secret  = local.secrets["beta"]
                    }
                }
            ]
        }
    }
}`

	rule := rules.NewSortAdditionalSecretsDataRule()
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
	alphaIdx := indexOf(fixed, "alpha-secret")
	betaIdx := indexOf(fixed, "beta-secret")
	zebraIdx := indexOf(fixed, "zebra-secret")

	if alphaIdx == -1 || betaIdx == -1 || zebraIdx == -1 {
		t.Fatalf("missing expected keys in fixed content: %s", fixed)
	}

	if !(alphaIdx < betaIdx && betaIdx < zebraIdx) {
		t.Errorf("keys not properly sorted. alpha=%d, beta=%d, zebra=%d\nfixed:\n%s",
			alphaIdx, betaIdx, zebraIdx, fixed)
	}
}

func TestSortAdditionalSecretsDataRule_RealWorldExample(t *testing.T) {
	content := `
locals {
    provider_to_app_map = {
        provider_a = {
            user_rails = {
                additional_parameters = []
                additional_secrets = [
                    {
                        name = "user-rails-master-key-secret"
                        data = {
                            codle-react-app-secret     = local.aws_secrets["codle_react_app_secret"]
                            google-client-secret       = local.aws_secrets["codle_google_client_secret"]
                            jupyter-hub-app-secret     = local.aws_secrets["jupyter_hub_app_secret"]
                            redis-password             = local.aws_secrets["elasticache_valkey_token"]
                            user-master-key            = local.aws_secrets["user_rails_master_key"]
                            user-secret-key-base       = local.aws_secrets["user_secret_key_base"]
                            smtp-password              = local.aws_secrets["smtp_password"]
                            teacherville-client-secret = local.aws_secrets["teacherville_client_secret"]
                            whalespace-client-secret   = local.aws_secrets["whalesapce_client_secret"]
                            microsoft-client-secret    = local.aws_secrets["microsoft_client_secret"]
                            whalespace-api-key         = local.aws_secrets["whalespace_api_key"]
                            naver-works-client-secret  = local.aws_secrets["naver_works_client_secret"]
                        }
                    }
                ]
            }
        }
    }
}`

	rule := rules.NewSortAdditionalSecretsDataRule()
	runner := helper.TestRunner(t, map[string]string{
		"local_config.tf": content,
	})

	if err := rule.Check(runner); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// Should trigger because keys are not sorted (redis-password should come after naver-works-client-secret, etc.)
	if len(runner.Issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(runner.Issues))
	}

	changes := runner.Changes()
	fixed := string(changes["local_config.tf"])

	// Verify sorting: codle < google < jupyter < microsoft < naver < redis < smtp < teacherville < user-master < user-secret < whalespace-api < whalespace-client
	codleIdx := indexOf(fixed, "codle-react-app-secret")
	googleIdx := indexOf(fixed, "google-client-secret")
	microsoftIdx := indexOf(fixed, "microsoft-client-secret")
	naverIdx := indexOf(fixed, "naver-works-client-secret")
	redisIdx := indexOf(fixed, "redis-password")

	if !(codleIdx < googleIdx && googleIdx < microsoftIdx && microsoftIdx < naverIdx && naverIdx < redisIdx) {
		t.Errorf("keys not properly sorted in real-world example\nfixed:\n%s", fixed)
	}
}
