package model

import (
	"errors"
	"strings"
)

var (
	ErrInvalidPhase = errors.New("invalid permit phase")
	ErrMissingID    = errors.New("missing identity")
)

func (p Permit) Valid() error {
	if strings.TrimSpace(p.ID) == "" || strings.TrimSpace(p.Worksite) == "" {
		return ErrMissingID
	}
	if p.Phase == "" {
		return ErrInvalidPhase
	}
	return nil
}

func (c Command) Complete() bool { return c.State == CommandConfirmed }

func (c Certificate) Current(nowUnix int64) bool {
	return c.Passed && nowUnix >= c.IssuedAt.Unix() && nowUnix <= c.ExpiresAt.Unix()
}

func (e TelemetryEvidence) SafeToRelease() bool {
	return e.Durable && e.ZeroCurrent && len(e.Samples) > 0
}
