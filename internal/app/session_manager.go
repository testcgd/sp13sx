package app

import (
	"sort"

	"sp13sx/internal/domain"
	"sp13sx/internal/store"
	"sp13sx/internal/util"
)

type SessionManager struct {
	paths store.Paths
}

func NewSessionManager(paths store.Paths) *SessionManager {
	return &SessionManager{paths: paths}
}

func (m *SessionManager) EnsureSession(backend string, model string) (domain.Session, error) {
	sessions, err := store.LoadSessions(m.paths.SessionsPath)
	if err != nil {
		return domain.Session{}, err
	}
	if len(sessions) > 0 {
		sort.Slice(sessions, func(i, j int) bool {
			return sessions[i].CreatedAt.After(sessions[j].CreatedAt)
		})
		return sessions[0], nil
	}
	now := util.NowUTC()
	session := domain.Session{
		ID:        util.NewID("sess"),
		Title:     "New Session",
		Backend:   backend,
		Model:     model,
		CreatedAt: now,
	}
	err = store.SaveSession(m.paths.SessionsPath, store.SessionRecord{
		ID:        util.NewID("evt"),
		Type:      "session.created",
		CreatedAt: now.Format(timeLayout),
		Payload:   session,
	})
	return session, err
}
