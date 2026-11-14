package routes

import (
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"

	"github.com/gofiber/fiber/v3"
)

// Legacy workflows are read-only for the compatibility release: interactions
// replaced them, and the write routes were removed with the old editor.
func WorkflowRoutes(group fiber.Router, repos *repositories.RepositoriesInterface) {
	eventWorkflowService := services.NewEventWorkflowService(repos.EventWorkflowRepository, repos.ProjectRepository)
	elementEventWorkflowService := services.NewElementEventWorkflowService(repos.ElementEventWorkflowRepository, repos.ElementRepository, repos.ProjectRepository, repos.PageRepository)
	projectService := services.NewProjectService(repos.ProjectRepository, repos.CollaboratorRepository, repos.UserRepository)
	pageService := services.NewPageService(repos.PageRepository, repos.ProjectRepository)

	eventWorkflowHandler := handlers.NewEventWorkflowHandler(eventWorkflowService, elementEventWorkflowService)
	elementEventWorkflowHandler := handlers.NewElementEventWorkflowHandler(elementEventWorkflowService, projectService, pageService)

	// Project-scoped list. The old GET /event-workflows/:projectid shadowed
	// GET /event-workflows/:id (same pattern, registered first), so it is gone.
	group.Get("/projects/:projectid/event-workflows", eventWorkflowHandler.GetEventWorkflowsByProject)
	group.Get("/event-workflows/:id", eventWorkflowHandler.GetEventWorkflowByID)
	group.Get("/event-workflows/:id/elements", eventWorkflowHandler.GetEventWorkflowElements)

	// Registered before /:id so "page" is not read as a workflow id.
	group.Get("/element-event-workflows/page/:pageId", elementEventWorkflowHandler.GetElementEventWorkflowsByPageId)
	group.Get("/element-event-workflows/:id", elementEventWorkflowHandler.GetElementEventWorkflowByID)
	group.Get("/element-event-workflows", elementEventWorkflowHandler.GetElementEventWorkflows)
}
