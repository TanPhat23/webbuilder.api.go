package routes

import (
	"my-go-app/internal/handlers"
	"my-go-app/internal/services"

	"github.com/gofiber/fiber/v3"
)

func InvitationRoutes(group fiber.Router, invitationService *services.InvitationService) {
	invitationHandler := handlers.NewInvitationHandler(invitationService)

	group.Post("/invitations", invitationHandler.CreateInvitation)
	group.Post("/invitations/accept", invitationHandler.AcceptInvitation)
	group.Post("/invitations/:token/accept", invitationHandler.AcceptInvitation)
	group.Get("/invitations", invitationHandler.GetInvitationsByProject)
	group.Get("/invitations/project/:projectid", invitationHandler.GetInvitationsByProject)
	group.Get("/invitations/project/:projectid/pending", invitationHandler.GetPendingInvitationsByProject)
	group.Patch("/invitations/:invitationid/cancel", invitationHandler.CancelInvitation)
	group.Patch("/invitations/:invitationid/status", invitationHandler.UpdateInvitationStatus)
	group.Delete("/invitations/:id", invitationHandler.DeleteInvitation)
	group.Delete("/invitations/:invitationid", invitationHandler.DeleteInvitation)
}
