package skills

import (
	"strings"

	"sp13sx/internal/domain"
)

func Compose(base string, enabled []string, registry *Registry) string {
	var parts []string
	if strings.TrimSpace(base) != "" {
		parts = append(parts, strings.TrimSpace(base))
	}
	for _, name := range enabled {
		skill, ok := registry.Get(name)
		if !ok {
			continue
		}
		parts = append(parts, renderSkill(skill))
	}
	return strings.TrimSpace(strings.Join(parts, "\n\n"))
}

func renderSkill(skill domain.Skill) string {
	if skill.Body == "" {
		return ""
	}
	return "# Skill: " + skill.Name + "\n\n" + skill.Body
}
