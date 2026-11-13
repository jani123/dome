package knowledge_test

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/jani123/dome/internal/knowledge"
	"github.com/stretchr/testify/require"
)

func memory(files map[string]string) fstest.MapFS {
	fs := fstest.MapFS{}
	for name, body := range files {
		fs[name] = &fstest.MapFile{Data: []byte(body)}
	}
	return fs
}

func validate(t *testing.T, files map[string]string, policy bool) []knowledge.Issue {
	t.Helper()
	issues, err := knowledge.Validate(memory(files), knowledge.Options{Bundle: "knowledge", Policy: policy})
	require.NoError(t, err)
	return issues
}

func TestMinimalExtensibleConformance(t *testing.T) {
	for _, body := range []string{
		"---\ntype: Future Custom Type\nunknown: [1, 2]\n---\nA minimal concept.",
		"---\ntype: Fact\n---\n[Missing](/missing.md)",
		"---\r\ntype: Fact\r\n---\r\nBody",
	} {
		require.Empty(t, validate(t, map[string]string{"knowledge/one.md": body}, false))
	}
	require.Empty(t, validate(t, map[string]string{"knowledge/index.md": "---\nokf_version: \"0.2\"\n---\n# Bundle\n"}, false))
}

func TestInvalidFrontmatter(t *testing.T) {
	for _, body := range []string{"No front matter", "---\ntype: [\n---\nbody", "---\ntype: Fact\n", "---\n- list\n---\nbody", "---\ntype: ''\n---\nbody", "---\ntype: 123\n---\nbody", "---\ntype: Fact\n---\n" + string([]byte{0xff}), "---\ntype: Fact\ntype: Duplicate\n---\nbody"} {
		issues := validate(t, map[string]string{"knowledge/one.md": body}, false)
		require.NotEmpty(t, issues)
		require.Equal(t, "format", issues[0].Rule)
	}
}

func TestReservedStructure(t *testing.T) {
	for name, body := range map[string]string{
		"knowledge/child/index.md": "---\ntype: Fact\n---\n# Wrong",
		"knowledge/index.md":       "---\nokf_version: \"0.2\"\ntype: Fact\n---\n# Wrong",
		"knowledge/log.md":         "# Log\n## yesterday\n- Changed\n",
	} {
		require.NotEmpty(t, validate(t, map[string]string{name: body}, false))
	}
	require.Empty(t, validate(t, map[string]string{"knowledge/log.md": "# Log\n## 2026-11-13\n- Updated\n## 2026-11-12\n- Created\n"}, false))
	require.NotEmpty(t, validate(t, map[string]string{"knowledge/log.md": "# Log\n## 2026-11-12\n- Updated\n## 2026-11-13\n- Created\n"}, false))
}

const concept = `---
type: Decision
title: Source access
description: Customers can rebuild agents.
generated: {by: codex/2026-11-13, at: '2026-11-13T07:00:00Z'}
sources:
  - {id: license, resource: ../LICENSE}
---
Customers can rebuild.[^license]

[^license]: License text
`

func fixture() map[string]string {
	return map[string]string{
		"knowledge/index.md":  "---\nokf_version: \"0.2\"\n---\n# Decisions\n- [Source access](source.md) - Customers can rebuild agents.\n",
		"knowledge/source.md": concept,
		"LICENSE":             "GPLv3 text",
	}
}

func TestPolicyAndEvidence(t *testing.T) {
	require.Empty(t, validate(t, fixture(), true))
	files := fixture()
	files["knowledge/source.md"] = strings.Replace(concept, "../LICENSE", "../missing", 1)
	issues := validate(t, files, true)
	require.NotEmpty(t, issues)
	require.Equal(t, "policy", issues[0].Rule)
	files = fixture()
	files["knowledge/source.md"] = strings.Replace(concept, "[^license]", "[^unknown]", 1)
	require.NotEmpty(t, validate(t, files, true))
	files = fixture()
	files["knowledge/source.md"] = strings.Replace(concept, "- {id: license, resource: ../LICENSE}", "- {id: license}", 1)
	require.NotEmpty(t, validate(t, files, true))
	files = fixture()
	files["knowledge/index.md"] = "# Index\n"
	require.NotEmpty(t, validate(t, files, true))
	files = fixture()
	files["knowledge/index.md"] = strings.Replace(files["knowledge/index.md"], "Customers can rebuild agents.", "Stale description.", 1)
	require.NotEmpty(t, validate(t, files, true))
}

func TestVerificationMetadata(t *testing.T) {
	for _, verification := range []string{
		"verified: {by: 'process:source-check', at: '2026-11-13T07:01:00Z'}\n",
		"verified:\n  - {by: 'human:reviewer', at: '2026-11-13T07:01:00Z'}\n",
	} {
		files := fixture()
		files["knowledge/source.md"] = strings.Replace(concept, "sources:", verification+"sources:", 1)
		require.Empty(t, validate(t, files, true))
	}
	for _, verification := range []string{
		"verified: invalid\n", "verified: {by: reviewer, at: invalid}\n",
		"verified: {by: 'process:source-check', at: '2026-11-12T07:00:00Z'}\n",
	} {
		files := fixture()
		files["knowledge/source.md"] = strings.Replace(concept, "sources:", verification+"sources:", 1)
		require.NotEmpty(t, validate(t, files, true))
	}
}

func TestBundleRelativeLinksAndEscape(t *testing.T) {
	files := fixture()
	files["knowledge/source.md"] += "\n[Self](/source.md) [Web](https://example.com/a)\n"
	require.Empty(t, validate(t, files, true))
	files["knowledge/source.md"] += "\n[Escape](../../outside.md)\n"
	require.NotEmpty(t, validate(t, files, true))
	_, err := knowledge.Validate(memory(files), knowledge.Options{Bundle: "../escape"})
	require.Error(t, err)
	_, err = knowledge.Validate(memory(nil), knowledge.Options{Bundle: "knowledge"})
	require.Error(t, err)
}

func TestDiagnosticsDoNotEchoContents(t *testing.T) {
	issues := validate(t, map[string]string{"knowledge/one.md": "---\ntype: [secret-token\n---\n"}, false)
	require.NotEmpty(t, issues)
	for _, issue := range issues {
		require.NotContains(t, issue.String(), "secret-token")
	}
}

func TestExamplesAreNotNavigationOrClaims(t *testing.T) {
	files := fixture()
	files["knowledge/source.md"] += "\n```markdown\n[Example](/absent.md) [^absent]\n```\n\n`[Example](/also-absent.md)`\n"
	require.Empty(t, validate(t, files, true))
	files["knowledge/source.md"] += "\n[Actual](/absent.md)\n"
	require.NotEmpty(t, validate(t, files, true))
}

func TestNestedDiscoverabilityAndSourceDirectories(t *testing.T) {
	files := fixture()
	delete(files, "knowledge/source.md")
	files["knowledge/decisions/source.md"] = strings.Replace(concept, "../LICENSE", "../../source-dir", 1)
	files["source-dir/evidence.go"] = "package evidence"
	files["knowledge/decisions/index.md"] = "# Decisions\n- [Source](source.md) - Customers can rebuild agents.\n"
	files["knowledge/index.md"] = "# Knowledge\n- [Decisions](decisions/)\n"
	require.Empty(t, validate(t, files, true))
	files["knowledge/index.md"] = "# Knowledge\n"
	require.NotEmpty(t, validate(t, files, true))
}
