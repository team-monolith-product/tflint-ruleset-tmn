package main

import (
	"github.com/team-monolith-product/tflint-ruleset-tmn/rules"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		RuleSet: &tflint.BuiltinRuleSet{
			Name:    "tmn",
			Version: "0.3.1",
			Rules: []tflint.Rule{
				rules.NewForEachTosetRule(),
				rules.NewSortProviderAppMapRule(),
				rules.NewSortAdditionalSecretsDataRule(),
				rules.NewSortAdditionalParametersDataRule(),
			},
		},
	})
}
