package genshell

import (
	"sort"

	"github.com/mgmaster24/config-gen-tools/shellforge/internal/config"
)

// Block is one integration's rendered contribution, already placed in its
// phase. Templates just walk Blocks in order — all ordering decisions are
// made here so they're testable without rendering anything.
type Block struct {
	Name    string
	Phase   string
	Guard   string
	Lines   []string
	Guarded bool
}

// TemplateData is everything the embedded template renders from. It contains
// no non-deterministic content, which the golden tests depend on.
type TemplateData struct {
	Shell  string
	Blocks []Block
}

// BuildTemplateData resolves cfg into an ordered, deterministic block list.
//
// Ordering is (phase, integration name) — never selection order. Two configs
// listing the same integrations in different orders must produce identical
// output, because the whole point is that shellforge decides placement
// rather than the user.
func BuildTemplateData(cfg config.Config) TemplateData {
	type entry struct {
		integration config.Integration
		spec        IntegrationSpec
	}

	var entries []entry
	seen := make(map[config.Integration]bool, len(cfg.Integrations))
	for _, i := range cfg.Integrations {
		if seen[i] {
			continue
		}
		seen[i] = true
		spec, ok := IntegrationSpecs[i]
		if !ok {
			continue
		}
		// An integration with no snippet for this shell is skipped rather
		// than emitting something that won't work.
		if len(spec.Snippets[cfg.Shell]) == 0 {
			continue
		}
		entries = append(entries, entry{i, spec})
	}

	sort.Slice(entries, func(a, b int) bool {
		if entries[a].spec.Phase != entries[b].spec.Phase {
			return entries[a].spec.Phase < entries[b].spec.Phase
		}
		return entries[a].integration < entries[b].integration
	})

	data := TemplateData{Shell: string(cfg.Shell)}
	for _, e := range entries {
		data.Blocks = append(data.Blocks, Block{
			Name:    e.integration.DisplayName(),
			Phase:   e.spec.Phase.String(),
			Guard:   e.spec.Binary,
			Lines:   e.spec.Snippets[cfg.Shell],
			Guarded: e.spec.Guarded,
		})
	}
	return data
}
