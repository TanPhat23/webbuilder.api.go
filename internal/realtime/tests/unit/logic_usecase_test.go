package unit

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	domainErrors "my-go-app/internal/realtime/domain/errors"
	"my-go-app/internal/realtime/domain/logic"
	"my-go-app/internal/realtime/usecase"
)

type failingLogicRepo struct{ err error }

func (r failingLogicRepo) Apply(context.Context, string, string, string, map[string]any) (logic.Revisioned, error) {
	return logic.Revisioned{}, r.err
}
func (r failingLogicRepo) Get(context.Context, string, string, string) (logic.Revisioned, logic.Revisioned, error) {
	return logic.Revisioned{}, logic.Revisioned{}, r.err
}
func (r failingLogicRepo) Access(context.Context, string, string) (bool, bool, error) {
	return false, false, r.err
}

func TestLogicUseCaseGivesRepositoryErrorsClientCodes(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		code string
	}{
		{"no permission", domainErrors.ErrUnauthorized, domainErrors.CodeAccessDenied},
		{"missing page", fmt.Errorf("%w: page not found", domainErrors.ErrNotFound), domainErrors.CodeNotFound},
		{"missing row", sql.ErrNoRows, domainErrors.CodeNotFound},
		{"invalid document", fmt.Errorf("%w: duplicate variable name", domainErrors.ErrValidation), domainErrors.CodeValidation},
		{"database failure", errors.New("pq: connection to 10.0.0.5 refused"), domainErrors.CodeInternal},
	} {
		uc := usecase.NewLogicUseCase(failingLogicRepo{tc.err})
		_, updateErr := uc.Update(context.Background(), "p", "pg", "u", map[string]any{})
		_, _, getErr := uc.Get(context.Background(), "p", "pg", "u")
		for op, err := range map[string]error{"update": updateErr, "get": getErr} {
			if !domainErrors.IsCode(err, tc.code) {
				t.Errorf("%s/%s: got %v, want code %s", tc.name, op, err, tc.code)
			}
		}
	}
	// A database failure must not leak its message to the client.
	_, err := usecase.NewLogicUseCase(failingLogicRepo{errors.New("pq: connection to 10.0.0.5 refused")}).Update(context.Background(), "p", "", "u", nil)
	var appErr *domainErrors.AppError
	if !errors.As(err, &appErr) || appErr.Message != "could not update logic" {
		t.Fatalf("internal errors must carry a generic message, got %v", err)
	}
}
