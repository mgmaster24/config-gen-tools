package genshell

import (
	"reflect"
	"testing"

	"github.com/mgmaster24/config-gen-tools/shellforge/internal/config"
)

func build(shell config.Shell, integrations ...config.Integration) TemplateData {
	return BuildTemplateData(config.Config{Shell: shell, Integrations: integrations})
}

func names(data TemplateData) []string {
	out := make([]string, 0, len(data.Blocks))
	for _, b := range data.Blocks {
		out = append(out, b.Name)
	}
	return out
}

// The reason shellforge exists: zoxide's own docs require its init to be the
// last thing evaluated, and it warns on every shell startup when it isn't.
// No selection or ordering of other integrations may displace it.
func TestBuildTemplateData_ZoxideIsAlwaysLast(t *testing.T) {
	selections := [][]config.Integration{
		{config.IntegrationZoxide},
		{config.IntegrationZoxide, config.IntegrationStarship},
		{config.IntegrationStarship, config.IntegrationZoxide},
		config.AllIntegrations,
		{config.IntegrationZoxide, config.IntegrationMise, config.IntegrationFzf},
	}

	for _, sel := range selections {
		for _, shell := range config.AllShells {
			data := build(shell, sel...)
			if len(data.Blocks) == 0 {
				t.Fatalf("no blocks for %v on %s", sel, shell)
			}
			last := data.Blocks[len(data.Blocks)-1]
			if last.Name != "zoxide" {
				t.Errorf("%s with %v: last block = %q, want zoxide (order: %v)",
					shell, sel, last.Name, names(data))
			}
		}
	}
}

// PATH-mutating integrations must precede everything that resolves binaries,
// or the prompt and tool hooks see a stale PATH.
func TestBuildTemplateData_PathPhasePrecedesToolAndPrompt(t *testing.T) {
	data := build(config.ShellZsh, config.AllIntegrations...)

	phaseOf := map[string]string{}
	order := make([]string, 0, len(data.Blocks))
	for _, b := range data.Blocks {
		phaseOf[b.Name] = b.Phase
		order = append(order, b.Phase)
	}

	rank := map[string]int{"path": 0, "tool": 1, "prompt": 2, "last": 3}
	for i := 1; i < len(order); i++ {
		if rank[order[i-1]] > rank[order[i]] {
			t.Fatalf("phases out of order at %d: %v", i, order)
		}
	}
	if phaseOf["mise"] != "path" {
		t.Errorf("mise phase = %q, want path", phaseOf["mise"])
	}
	if phaseOf["starship"] != "prompt" {
		t.Errorf("starship phase = %q, want prompt", phaseOf["starship"])
	}
}

// Placement is shellforge's decision, not the user's: the same set selected
// in a different order must produce byte-identical output.
func TestBuildTemplateData_IsOrderIndependent(t *testing.T) {
	a := build(config.ShellZsh, config.IntegrationZoxide, config.IntegrationFzf, config.IntegrationMise)
	b := build(config.ShellZsh, config.IntegrationMise, config.IntegrationZoxide, config.IntegrationFzf)

	if !reflect.DeepEqual(a, b) {
		t.Errorf("selection order leaked into output:\na = %v\nb = %v", names(a), names(b))
	}
}

func TestBuildTemplateData_DedupesRepeatedIntegration(t *testing.T) {
	data := build(config.ShellZsh, config.IntegrationFzf, config.IntegrationFzf)
	if len(data.Blocks) != 1 {
		t.Errorf("got %d blocks, want 1: %v", len(data.Blocks), names(data))
	}
}

func TestBuildTemplateData_UnselectedIntegrationLeaksNothing(t *testing.T) {
	data := build(config.ShellZsh, config.IntegrationFzf)
	for _, b := range data.Blocks {
		if b.Name != "fzf" {
			t.Errorf("block %q present though only fzf was selected", b.Name)
		}
	}
}

// Every integration must have a snippet for every supported shell, or a user
// selecting it on that shell silently gets nothing.
func TestIntegrationSpecs_CoverEveryShell(t *testing.T) {
	for _, i := range config.AllIntegrations {
		spec, ok := IntegrationSpecs[i]
		if !ok {
			t.Errorf("no IntegrationSpec for %q", i)
			continue
		}
		if spec.Binary == "" {
			t.Errorf("%q has no Binary set", i)
		}
		for _, shell := range config.AllShells {
			if len(spec.Snippets[shell]) == 0 {
				t.Errorf("%q has no snippet for %s", i, shell)
			}
		}
	}
}
