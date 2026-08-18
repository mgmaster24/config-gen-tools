package checks

import (
	"testing"

	"github.com/mgmaster24/config-gen-tools/forge/prereq"

	"github.com/mgmaster24/config-gen-tools/nvimforge/internal/config"
)

func TestLanguageChecks_CoversEveryLanguage(t *testing.T) {
	for _, l := range config.AllLanguages {
		if _, ok := LanguageChecks[l]; !ok {
			t.Errorf("LanguageChecks has no entry for %q (add one, even if empty)", l)
		}
	}
}

func TestChecks_HaveNameSeverityAndAtLeastOneDetectionMethod(t *testing.T) {
	all := append([]prereq.Check{}, UniversalChecks...)
	for _, checks := range LanguageChecks {
		all = append(all, checks...)
	}

	for _, c := range all {
		if c.Name == "" {
			t.Errorf("check with description %q has empty Name", c.Description)
		}
		if c.Description == "" {
			t.Errorf("check %q has empty Description", c.Name)
		}
		if c.Binary == "" && c.Detect == nil {
			t.Errorf("check %q has neither Binary nor Detect set", c.Name)
		}
		if c.Severity != prereq.SeverityRecommended && c.Severity != prereq.SeverityRequired {
			t.Errorf("check %q has unrecognized Severity %v", c.Name, c.Severity)
		}
	}
}
