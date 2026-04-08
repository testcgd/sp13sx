package skills

import (
	"strings"

	"gopkg.in/yaml.v3"
)

type Frontmatter struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Metadata    map[string]string `yaml:"metadata"`
}

func ParseSkillDocument(input string) (Frontmatter, string, error) {
	if !strings.HasPrefix(input, "---\n") {
		return Frontmatter{}, strings.TrimSpace(input), nil
	}
	rest := strings.TrimPrefix(input, "---\n")
	idx := strings.Index(rest, "\n---\n")
	if idx < 0 {
		return Frontmatter{}, strings.TrimSpace(input), nil
	}

	rawMeta := rest[:idx]
	body := strings.TrimSpace(rest[idx+5:])

	var fm Frontmatter
	if err := yaml.Unmarshal([]byte(rawMeta), &fm); err != nil {
		return Frontmatter{}, "", err
	}
	return fm, body, nil
}
