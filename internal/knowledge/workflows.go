package knowledge

import (
	"io/fs"
	"strings"

	"github.com/jani123/dome/core/fault"
)

// ValidateWorkflows checks that required maintainer skills match the pinned CLI.
// Other independently installed skills are deliberately outside this check.
func ValidateWorkflows(repo fs.FS) ([]Issue, error) {
	pin, err := fs.ReadFile(repo, ".openspec-version")
	if err != nil {
		return nil, fault.Wrap(fault.Configuration, err)
	}
	version := strings.TrimSpace(string(pin))
	issues := []Issue{}
	if version == "" {
		return []Issue{{".openspec-version", "policy", "OpenSpec version pin must not be empty"}}, nil
	}
	for _, name := range []string{"openspec-explore", "openspec-propose", "openspec-apply-change", "openspec-update-change", "openspec-sync-specs", "openspec-archive-change", "openspec-verify-change"} {
		file := ".agents/skills/" + name + "/SKILL.md"
		body, err := fs.ReadFile(repo, file)
		if err != nil {
			return nil, fault.Wrap(fault.Configuration, err)
		}
		doc, err := parse(body)
		if err != nil || text(doc.meta["name"]) != name || text(doc.meta["description"]) == "" || text(mapping(doc.meta["metadata"])["generatedBy"]) != version {
			issues = append(issues, Issue{file, "policy", "workflow metadata must identify its skill and pinned OpenSpec generator version"})
		}
	}
	return issues, nil
}
