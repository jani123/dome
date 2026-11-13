// Package knowledge validates developer documentation, not consumer runtime state.
// Format checks target Google OKF v0.2; Policy adds Dome authoring conventions.
package knowledge

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jani123/dome/core/fault"
	"gopkg.in/yaml.v3"
)

// Issue distinguishes an upstream format error from a repository-policy finding.
type Issue struct{ Path, Rule, Message string }

func (i Issue) String() string { return fmt.Sprintf("%s [%s]: %s", i.Path, i.Rule, i.Message) }

// Options selects the repository-relative bundle and optional Dome policy.
type Options struct {
	Bundle string
	Policy bool
}

type document struct {
	meta     map[string]any
	body     string
	hasFront bool
}

var (
	heading      = regexp.MustCompile(`(?m)^#{1,6} +\S`)
	datedHeading = regexp.MustCompile(`(?m)^## +([^\r\n]+)`)
	links        = regexp.MustCompile(`\[[^\]\n]*\]\(<?([^\s)>]+)>?(?: +"[^"\n]*")?\)`)
	footnotes    = regexp.MustCompile(`\[\^([^\]\n]+)\]`)
	actor        = regexp.MustCompile(`^(human:[^\s]+|process:[^\s]+|[^\s/:]+/[^\s]+)$`)
)

// Validate reads only the injected filesystem. It never follows network links,
// rewrites documents, or includes document contents in diagnostics.
func Validate(repo fs.FS, options Options) ([]Issue, error) {
	bundle := path.Clean(options.Bundle)
	if !fs.ValidPath(bundle) {
		return nil, fault.New(fault.Validation, "bundle must be a repository-relative path")
	}
	issues := []Issue{}
	add := func(file, rule, message string) { issues = append(issues, Issue{file, rule, message}) }
	docs := map[string]document{}
	err := fs.WalkDir(repo, bundle, func(file string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if options.Policy && entry.Type()&fs.ModeSymlink != 0 {
			add(file, "policy", "symlinks are not supported in a knowledge bundle")
			return nil
		}
		if entry.IsDir() || !strings.HasSuffix(file, ".md") {
			return nil
		}
		data, err := fs.ReadFile(repo, file)
		if err != nil {
			return err
		}
		doc, err := parse(data)
		if err != nil {
			add(file, "format", "invalid UTF-8, YAML mapping, or front-matter delimiters")
			return nil
		}
		docs[file] = doc
		base := path.Base(file)
		switch base {
		case "index.md":
			if doc.hasFront && (file != path.Join(bundle, "index.md") || len(doc.meta) != 1 || doc.meta["okf_version"] != "0.2") {
				add(file, "format", "only the bundle-root index may carry front matter containing okf_version: \"0.2\"")
			}
			if !heading.MatchString(prose(doc.body)) {
				add(file, "format", "index needs a section heading")
			}
		case "log.md":
			if doc.hasFront {
				add(file, "format", "reserved log files have no front matter")
			}
			dates := datedHeading.FindAllStringSubmatch(prose(doc.body), -1)
			if len(dates) == 0 {
				add(file, "format", "log needs ISO-date section headings")
			}
			previous := "9999-12-31"
			for _, m := range dates {
				date := strings.TrimSpace(m[1])
				if _, err := time.Parse(time.DateOnly, date); err != nil || date > previous {
					add(file, "format", "log date headings must use YYYY-MM-DD in newest-first order")
				}
				previous = date
			}
		default:
			if text(doc.meta["type"]) == "" {
				add(file, "format", "concept requires a non-empty string type")
			}
		}
		return nil
	})
	if err != nil {
		return nil, fault.Wrap(fault.Configuration, err)
	}
	if options.Policy {
		if _, ok := docs[path.Join(bundle, "index.md")]; !ok {
			add(bundle, "policy", "bundle-root index is required")
		}
		names := make([]string, 0, len(docs))
		for file := range docs {
			names = append(names, file)
		}
		sort.Strings(names)
		for _, file := range names {
			doc := docs[file]
			base := path.Base(file)
			if base != "index.md" && base != "log.md" {
				checkConcept(repo, bundle, file, doc, add)
				parent := path.Join(path.Dir(file), "index.md")
				index, ok := docs[parent]
				indexed := false
				for _, m := range links.FindAllStringSubmatch(prose(index.body), -1) {
					target, local, err := resolve(bundle, parent, m[1])
					if err == nil && local && target == file {
						indexed = true
					}
				}
				if !ok || !indexed {
					add(file, "policy", "concept must be linked from its directory index")
				}
			}
			for _, m := range links.FindAllStringSubmatch(prose(doc.body), -1) {
				checkReference(repo, bundle, file, m[1], add)
			}
			if base == "index.md" {
				if file != path.Join(bundle, "index.md") {
					parent := path.Join(path.Dir(path.Dir(file)), "index.md")
					linked := false
					for _, m := range links.FindAllStringSubmatch(prose(docs[parent].body), -1) {
						target, local, err := resolve(bundle, parent, m[1])
						if err == nil && local && (target == file || target == path.Dir(file)) {
							linked = true
						}
					}
					if !linked {
						add(file, "policy", "directory index must be linked from its parent index")
					}
				}
				for _, m := range links.FindAllStringSubmatch(prose(doc.body), -1) {
					target, local, err := resolve(bundle, file, m[1])
					if err != nil || !local {
						continue
					}
					if strings.HasSuffix(m[1], "/") {
						target = path.Join(target, "index.md")
					}
					if child, ok := docs[target]; ok && path.Base(target) != "index.md" {
						// Index summaries are a search surface: prevent descriptions from drifting.
						for _, line := range strings.Split(prose(doc.body), "\n") {
							if strings.Contains(line, m[0]) && !strings.Contains(line, text(child.meta["description"])) {
								add(file, "policy", "index summary differs from linked concept description")
							}
						}
					}
				}
			}
		}
	}
	sort.Slice(issues, func(i, j int) bool { return issues[i].String() < issues[j].String() })
	return issues, nil
}

func parse(data []byte) (document, error) {
	if !utf8.Valid(data) {
		return document{}, fmt.Errorf("invalid encoding")
	}
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	doc := document{meta: map[string]any{}, body: string(data)}
	if !bytes.HasPrefix(data, []byte("---\n")) {
		return doc, nil
	}
	end := bytes.Index(data[4:], []byte("\n---\n"))
	if end < 0 && bytes.HasSuffix(data, []byte("\n---")) {
		data = append(data, '\n')
		end = bytes.Index(data[4:], []byte("\n---\n"))
	}
	if end < 0 {
		return doc, fmt.Errorf("unclosed front matter")
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data[4 : 4+end]))
	var node yaml.Node
	if err := decoder.Decode(&node); err != nil {
		return doc, err
	}
	if len(node.Content) != 1 || node.Content[0].Kind != yaml.MappingNode {
		return doc, fmt.Errorf("expected mapping")
	}
	if err := node.Decode(&doc.meta); err != nil {
		return doc, err
	}
	doc.meta = retainTimestampText(doc.meta, &node).(map[string]any)
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return doc, fmt.Errorf("multiple YAML documents")
	}
	if doc.meta == nil {
		return doc, fmt.Errorf("expected mapping")
	}
	doc.hasFront = true
	doc.body = string(data[4+end+5:])
	return doc, nil
}
func text(v any) string            { s, _ := v.(string); return strings.TrimSpace(s) }
func mapping(v any) map[string]any { m, _ := v.(map[string]any); return m }
func timestamp(v any) (time.Time, bool) {
	if t, ok := v.(time.Time); ok {
		return t, true
	}
	t, err := time.Parse(time.RFC3339, text(v))
	return t, err == nil
}

func checkConcept(repo fs.FS, bundle, file string, doc document, add func(string, string, string)) {
	for _, field := range []string{"title", "description"} {
		if text(doc.meta[field]) == "" {
			add(file, "policy", field+" is required for curated concepts")
		}
	}
	generated := mapping(doc.meta["generated"])
	generatedAt, timeOK := timestamp(generated["at"])
	if !actor.MatchString(text(generated["by"])) || !timeOK {
		add(file, "policy", "generation provenance needs an actor and an RFC3339 timestamp")
	}
	sources, ok := doc.meta["sources"].([]any)
	if !ok || len(sources) == 0 {
		add(file, "policy", "at least one evidence source is required")
	}
	ids := map[string]bool{}
	for _, s := range sources {
		source := mapping(s)
		resource := text(source["resource"])
		if resource == "" {
			add(file, "policy", "source requires a resource reference")
			continue
		}
		checkReference(repo, bundle, file, resource, add)
		if id := text(source["id"]); id != "" {
			if ids[id] {
				add(file, "policy", "source IDs must be unique")
			}
			ids[id] = true
		}
	}
	for _, m := range footnotes.FindAllStringSubmatch(prose(doc.body), -1) {
		if !ids[m[1]] {
			add(file, "policy", "claim footnote requires a matching source ID")
		}
	}
	if v, exists := doc.meta["verified"]; exists {
		events, ok := v.([]any)
		if !ok {
			if m := mapping(v); m != nil {
				events = []any{m}
			} else {
				add(file, "policy", "verification must be a mapping or list of events")
			}
		}
		for _, e := range events {
			m := mapping(e)
			at, ok := timestamp(m["at"])
			if !actor.MatchString(text(m["by"])) || !ok {
				add(file, "policy", "verification event needs an actor and RFC3339 timestamp")
			} else if timeOK && at.Before(generatedAt) {
				add(file, "policy", "verification predates the current content; recheck or remove it")
			}
		}
	}
	if v, exists := doc.meta["stale_after"]; exists {
		if _, ok := timestamp(v); !ok {
			add(file, "policy", "stale_after must be an RFC3339 timestamp")
		}
	}
}

func resolve(bundle, file, reference string) (string, bool, error) {
	u, err := url.Parse(reference)
	if err != nil {
		return "", true, err
	}
	if u.IsAbs() || u.Host != "" {
		return "", false, nil
	}
	if u.Path == "" {
		return file, true, nil
	}
	target := path.Join(path.Dir(file), u.Path)
	if strings.HasPrefix(u.Path, "/") {
		target = path.Join(bundle, strings.TrimPrefix(u.Path, "/"))
	}
	if !fs.ValidPath(target) {
		return "", true, fmt.Errorf("reference leaves repository")
	}
	return target, true, nil
}

func checkReference(repo fs.FS, bundle, file, reference string, add func(string, string, string)) {
	target, local, err := resolve(bundle, file, reference)
	if err != nil {
		add(file, "policy", "invalid local reference")
		return
	}
	if !local {
		return
	}
	info, err := fs.Stat(repo, target)
	if err != nil {
		add(file, "policy", "local reference does not exist: "+target)
		return
	}
	if info.IsDir() && (target == bundle || strings.HasPrefix(target, bundle+"/")) {
		if _, err := fs.Stat(repo, path.Join(target, "index.md")); err != nil {
			add(file, "policy", "linked knowledge directory needs an index")
		}
	}
}

// prose removes fenced examples and inline code before scanning actual links.
// It intentionally does not implement a renderer or validate heading anchors.
func prose(body string) string {
	lines := strings.Split(body, "\n")
	marker := byte(0)
	width := 0
	for i, line := range lines {
		trimmed := strings.TrimLeft(line, " ")
		if len(line)-len(trimmed) <= 3 && len(trimmed) >= 3 && (trimmed[0] == '`' || trimmed[0] == '~') {
			count := 0
			for count < len(trimmed) && trimmed[count] == trimmed[0] {
				count++
			}
			if marker == 0 && count >= 3 {
				marker, width = trimmed[0], count
				lines[i] = ""
				continue
			}
			if marker != 0 && trimmed[0] == marker && count >= width && strings.TrimSpace(trimmed[count:]) == "" {
				marker = 0
				lines[i] = ""
				continue
			}
		}
		if marker != 0 {
			lines[i] = ""
			continue
		}
		var out strings.Builder
		for pos := 0; pos < len(line); {
			if line[pos] != '`' {
				out.WriteByte(line[pos])
				pos++
				continue
			}
			end := pos
			for end < len(line) && line[end] == '`' {
				end++
			}
			delimiter := line[pos:end]
			close := strings.Index(line[end:], delimiter)
			if close < 0 {
				out.WriteString(delimiter)
				pos = end
				continue
			}
			pos = end + close + len(delimiter)
		}
		lines[i] = out.String()
	}
	return strings.Join(lines, "\n")
}

// YAML timestamp decoding otherwise loses whether the input had an offset (or
// even a time of day). Preserve original scalar text for explicit RFC3339 checks.
func retainTimestampText(value any, node *yaml.Node) any {
	switch node.Kind {
	case yaml.DocumentNode:
		return retainTimestampText(value, node.Content[0])
	case yaml.AliasNode:
		return retainTimestampText(value, node.Alias)
	case yaml.ScalarNode:
		if node.Tag == "!!timestamp" {
			return node.Value
		}
	case yaml.MappingNode:
		if m, ok := value.(map[string]any); ok {
			for i := 0; i+1 < len(node.Content); i += 2 {
				key := node.Content[i].Value
				m[key] = retainTimestampText(m[key], node.Content[i+1])
			}
		}
	case yaml.SequenceNode:
		if items, ok := value.([]any); ok {
			for i, item := range node.Content {
				items[i] = retainTimestampText(items[i], item)
			}
		}
	}
	return value
}
