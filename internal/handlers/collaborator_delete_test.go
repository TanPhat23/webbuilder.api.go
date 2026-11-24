package handlers

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"my-go-app/internal/services"
)

type deleteCollaboratorStub struct {
	services.CollaboratorServiceInterface
	deleted string
}

func (s *deleteCollaboratorStub) DeleteCollaborator(_ context.Context, id, _ string) error {
	s.deleted = id
	return nil
}

func TestDeleteCollaboratorReadsTheRegisteredParameter(t *testing.T) {
	stub := &deleteCollaboratorStub{}
	app := newKitApp("u1", func(app *fiber.App) {
		app.Delete("/collaborators/:collaboratorid", NewCollaboratorHandler(stub).DeleteCollaborator)
	})
	response, err := app.Test(httptest.NewRequest("DELETE", "/collaborators/collab-1", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != 204 || stub.deleted != "collab-1" {
		t.Fatalf("status=%d deleted=%q", response.StatusCode, stub.deleted)
	}
}
