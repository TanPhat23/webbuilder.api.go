package usecase

import (
	"context"
	"database/sql"
	"errors"

	domainErrors "my-go-app/internal/realtime/domain/errors"
	"my-go-app/internal/realtime/domain/logic"
	domainRepo "my-go-app/internal/realtime/domain/repository"
)

type logicUseCase struct{ repo domainRepo.LogicRepository }

func NewLogicUseCase(repo domainRepo.LogicRepository) *logicUseCase {
	return &logicUseCase{repo: repo}
}

// mapLogicError gives repository failures the codes clients understand.
// Anything unrecognised is an internal error, so database details are not leaked.
func mapLogicError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, domainErrors.ErrUnauthorized):
		return domainErrors.New(domainErrors.CodeAccessDenied, "you cannot edit this project's logic", err)
	case errors.Is(err, domainErrors.ErrNotFound), errors.Is(err, sql.ErrNoRows):
		return domainErrors.New(domainErrors.CodeNotFound, "project or page not found", err)
	case errors.Is(err, domainErrors.ErrValidation):
		return domainErrors.New(domainErrors.CodeValidation, err.Error(), err)
	default:
		return domainErrors.New(domainErrors.CodeInternal, "could not update logic", err)
	}
}

func (u *logicUseCase) Update(ctx context.Context, projectID, pageID, userID string, patch map[string]any) (logic.Revisioned, error) {
	doc, err := u.repo.Apply(ctx, projectID, pageID, userID, patch)
	if err != nil {
		return logic.Revisioned{}, mapLogicError(err)
	}
	return doc, nil
}

func (u *logicUseCase) Get(ctx context.Context, projectID, pageID, userID string) (logic.Revisioned, logic.Revisioned, error) {
	project, page, err := u.repo.Get(ctx, projectID, pageID, userID)
	if err != nil {
		return logic.Revisioned{}, logic.Revisioned{}, mapLogicError(err)
	}
	return project, page, nil
}

func (u *logicUseCase) Access(ctx context.Context, projectID, userID string) (bool, bool, error) {
	return u.repo.Access(ctx, projectID, userID)
}
