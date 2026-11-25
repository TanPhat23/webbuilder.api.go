package routes

import (
	"my-go-app/internal/handlers"
	"my-go-app/internal/services"

	"github.com/gofiber/fiber/v3"
)

func RegisterInvitations(group fiber.Router, invitationService *services.InvitationService) {
	invitationHandler := handlers.NewInvitationHandler(invitationService)

	group.Post("/invitations", invitationHandler.Create)
	group.Post("/invitations/accept", invitationHandler.Accept)
	group.Post("/invitations/:token/accept", invitationHandler.Accept)
	group.Get("/invitations", invitationHandler.ListByProject)
	group.Get("/invitations/project/:projectid", invitationHandler.ListByProject)
	group.Get("/invitations/project/:projectid/pending", invitationHandler.ListPendingByProject)
	group.Patch("/invitations/:invitationid/cancel", invitationHandler.Cancel)
	group.Patch("/invitations/:invitationid/status", invitationHandler.UpdateStatus)
	group.Delete("/invitations/:invitationid", invitationHandler.Delete)
}
