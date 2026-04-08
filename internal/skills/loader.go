package skills

import (
	"os"
	"path/filepath"

	"sp13sx/internal/domain"
)

func Discover(paths []string) ([]domain.Skill, error) {
	var out []domain.Skill
	for _, base := range paths {
		entries, err := os.ReadDir(base)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			skillPath := filepath.Join(base, entry.Name(), "SKILL.md")
			data, err := os.ReadFile(skillPath)
			if err != nil {
				continue
			}
			fm, body, err := ParseSkillDocument(string(data))
			if err != nil {
				return nil, err
			}
			name := entry.Name()
			if fm.Name != "" {
				name = fm.Name
			}
			out = append(out, domain.Skill{
				Name:        name,
				Description: fm.Description,
				Path:        filepath.Dir(skillPath),
				Metadata:    fm.Metadata,
				Body:        body,
			})
		}
	}
	return out, nil
}
