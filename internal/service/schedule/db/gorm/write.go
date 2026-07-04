package gorm

import (
	"context"
	"errors"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/schedule"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/shared"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/schedule/v1/schedulepbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
	"gorm.io/gorm"
)

// mapGormErr translates GORM sentinel errors into the provider-neutral errors in
// internal/types.
func mapGormErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return types.ErrNotFound
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return types.ErrAlreadyExists
	default:
		return err
	}
}

// UpdateSchedule upserts a unit's schedule configuration and returns the result.
// The schedule is a singleton (created on first update). An empty paths slice
// replaces every mutable section; s.Etag guards against concurrent writes on an
// existing schedule. The merged proto is re-materialized into a fresh child graph
// and the superseded rows are deleted in the same transaction.
func (r *ScheduleRepository) UpdateSchedule(ctx context.Context, s *schedulepbv1.Schedule, paths []string) (*schedulepbv1.Schedule, error) {
	propertyID, _, err := types.ParseScheduleName(s.GetName())
	if err != nil {
		return nil, err
	}
	err = r.db.Transaction(func(tx *gorm.DB) error {
		var existing schedule.Schedule
		loadErr := preloadSchedule(tx.WithContext(ctx)).First(&existing, "name = ?", s.GetName()).Error
		exists := loadErr == nil
		if loadErr != nil && !errors.Is(loadErr, gorm.ErrRecordNotFound) {
			return loadErr
		}
		if exists && s.GetEtag() != "" && existing.Etag != nil && s.GetEtag() != *existing.Etag {
			return types.ErrConflict
		}

		merged := &schedulepbv1.Schedule{Name: s.GetName()}
		if exists {
			merged = scheduleFromModel(&existing)
		}
		applyScheduleMask(merged, s, paths)
		g := buildScheduleGraph(merged, propertyID)

		// New belongs-to children.
		if g.buffers != nil {
			if e := schedule.NewBufferSettingsStore(tx).Create(ctx, g.buffers); e != nil {
				return e
			}
		}
		if g.stayConstraints != nil {
			if e := schedule.NewStayConstraintsStore(tx).Create(ctx, g.stayConstraints); e != nil {
				return e
			}
		}
		if g.cancellationPolicy != nil {
			if e := schedule.NewCancellationPolicyStore(tx).Create(ctx, g.cancellationPolicy); e != nil {
				return e
			}
		}

		if exists {
			oldBuffers, oldStay, oldCancel := existing.BuffersID, existing.StayConstraintsID, existing.CancellationPolicyID
			existing.BuffersID = g.schedule.BuffersID
			existing.StayConstraintsID = g.schedule.StayConstraintsID
			existing.CancellationPolicyID = g.schedule.CancellationPolicyID
			existing.Etag = ptr(ulid.GenerateString())
			existing.Buffers, existing.StayConstraints, existing.CancellationPolicy = nil, nil, nil
			existing.RecurringRules, existing.Exceptions = nil, nil
			if e := schedule.NewScheduleStore(tx).Update(ctx, &existing); e != nil {
				return e
			}
			g.schedule.ID = existing.ID
			if e := tx.WithContext(ctx).Where("schedule_id = ?", existing.ID).Delete(&schedule.RecurringRule{}).Error; e != nil {
				return e
			}
			if e := g.persistChildren(ctx, tx); e != nil {
				return e
			}
			return deleteScheduleChildren(ctx, tx, oldBuffers, oldStay, oldCancel)
		}

		g.schedule.ID = ulid.GenerateString()
		g.schedule.Name = s.GetName()
		g.schedule.Etag = ptr(ulid.GenerateString())
		if e := schedule.NewScheduleStore(tx).Create(ctx, g.schedule); e != nil {
			return e
		}
		return g.persistChildren(ctx, tx)
	})
	if err != nil {
		return nil, mapGormErr(err)
	}
	return r.GetSchedule(ctx, s.GetName())
}

// deleteScheduleChildren removes a schedule's superseded belongs-to rows once the
// schedule no longer references them: the cancellation policy (its refund tiers
// cascade in the DB), the buffer settings, and the stay constraints.
func deleteScheduleChildren(ctx context.Context, tx *gorm.DB, oldBuffers, oldStay, oldCancel *string) error {
	if oldCancel != nil {
		if e := schedule.NewCancellationPolicyStore(tx).DeleteByID(ctx, *oldCancel); e != nil {
			return e
		}
	}
	if oldBuffers != nil {
		if e := schedule.NewBufferSettingsStore(tx).DeleteByID(ctx, *oldBuffers); e != nil {
			return e
		}
	}
	if oldStay != nil {
		if e := schedule.NewStayConstraintsStore(tx).DeleteByID(ctx, *oldStay); e != nil {
			return e
		}
	}
	return nil
}

// DeleteAvailabilityException removes an exception and the TimeWindow / DateRange
// value-object its span referenced.
func (r *ScheduleRepository) DeleteAvailabilityException(ctx context.Context, name string) error {
	id, err := types.AvailabilityExceptionID(name)
	if err != nil {
		return err
	}
	err = r.db.Transaction(func(tx *gorm.DB) error {
		var existing schedule.AvailabilityException
		if e := tx.WithContext(ctx).First(&existing, "id = ?", id).Error; e != nil {
			return e
		}
		if e := schedule.NewAvailabilityExceptionStore(tx).DeleteByID(ctx, id); e != nil {
			return e
		}
		if existing.WindowID != nil {
			if e := shared.NewTimeWindowStore(tx).DeleteByID(ctx, *existing.WindowID); e != nil {
				return e
			}
		}
		if existing.DateRangeID != nil {
			if e := shared.NewDateRangeStore(tx).DeleteByID(ctx, *existing.DateRangeID); e != nil {
				return e
			}
		}
		return nil
	})
	return mapGormErr(err)
}
