package skills

import "sp13sx/internal/domain"

type Registry struct {
	skills map[string]domain.Skill
}

func NewRegistry(all []domain.Skill) *Registry {
	m := make(map[string]domain.Skill, len(all))
	for _, skill := range all {
		m[skill.Name] = skill
	}
	return &Registry{skills: m}
}

func (r *Registry) All() []domain.Skill {
	out := make([]domain.Skill, 0, len(r.skills))
	for _, skill := range r.skills {
		out = append(out, skill)
	}
	return out
}

func (r *Registry) Get(name string) (domain.Skill, bool) {
	skill, ok := r.skills[name]
	return skill, ok
}
