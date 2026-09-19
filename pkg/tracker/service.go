package tracker

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"codeberg.org/veya/ermokie/pkg/db"
)

var ErrNoActiveRun = errors.New("no active run")

type Service struct {
	store *db.Store
}

func New(store *db.Store) *Service {
	return &Service{store: store}
}

func (s *Service) requireActiveRun(ctx context.Context) error {
	if _, err := s.store.GetActiveRun(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNoActiveRun
		}
		return fmt.Errorf("get active run: %w", err)
	}
	return nil
}

func (s *Service) AddHit(ctx context.Context) error {
	if err := s.requireActiveRun(ctx); err != nil {
		return err
	}
	if err := s.store.IncrementActiveSplitHit(ctx); err != nil {
		return fmt.Errorf("add hit: %w", err)
	}
	return nil
}

func (s *Service) RemoveHit(ctx context.Context) error {
	if err := s.requireActiveRun(ctx); err != nil {
		return err
	}
	if err := s.store.DecrementActiveSplitHit(ctx); err != nil {
		return fmt.Errorf("remove hit: %w", err)
	}
	return nil
}

func (s *Service) Advance(ctx context.Context) error {
	if err := s.requireActiveRun(ctx); err != nil {
		return err
	}
	if err := s.store.AdvanceSplitInActiveRun(ctx); err != nil {
		return fmt.Errorf("advance split: %w", err)
	}
	return nil
}

func (s *Service) Previous(ctx context.Context) error {
	if err := s.requireActiveRun(ctx); err != nil {
		return err
	}
	if err := s.store.GoBackSplitInActiveRun(ctx); err != nil {
		return fmt.Errorf("previous split: %w", err)
	}
	return nil
}

func (s *Service) Reset(ctx context.Context) (err error) {
	if activeErr := s.requireActiveRun(ctx); activeErr != nil {
		return activeErr
	}
	tx, err := s.store.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin reset: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	queries := s.store.WithTx(tx)
	if err := queries.ResetActiveRunHits(ctx); err != nil {
		return fmt.Errorf("reset hits: %w", err)
	}
	if err := queries.ResetActiveRunProgress(ctx); err != nil {
		return fmt.Errorf("reset progress: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit reset: %w", err)
	}
	return nil
}
