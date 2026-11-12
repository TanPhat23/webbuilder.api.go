package routes

import (
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"

	"github.com/gofiber/fiber/v3"
)

func WorkflowRoutes(group fiber.Router, repos *repositories.RepositoriesInterface) {
	eventWorkflowService := services.NewEventWorkflowService(repos.EventWorkflowRepository, repos.ProjectRepository)
	elementEventWorkflowService := services.NewElementEventWorkflowService(repos.ElementEventWorkflowRepository, repos.ElementRepository, repos.ProjectRepository, repos.PageRepository)
	projectService := services.NewProjectService(repos.ProjectRepository, repos.CollaboratorRepository, repos.UserRepository)
	pageService := services.NewPageService(repos.PageRepository, repos.ProjectRepository)

	eventWorkflowHandler := handlers.NewEventWorkflowHandler(eventWorkflowService, elementEventWorkflowService)
	elementEventWorkflowHandler := handlers.NewElementEventWorkflowHandler(elementEventWorkflowService, projectService, pageService)

	group.Post("/event-workflows", eventWorkflowHandler.CreateEventWorkflow)
	group.Get("/event-workflows/:projectid", eventWorkflowHandler.GetEventWorkflowsByProject)
	group.Get("/event-workflows/:id", eventWorkflowHandler.GetEventWorkflowByID)
	group.Patch("/event-workflows/:id", eventWorkflowHandler.UpdateEventWorkflow)
	group.Patch("/event-workflows/:id/enabled", eventWorkflowHandler.UpdateEventWorkflowEnabled)
	group.Delete("/event-workflows/:id", eventWorkflowHandler.DeleteEventWorkflow)
	group.Get("/event-workflows/:id/elements", eventWorkflowHandler.GetEventWorkflowElements)

	group.Post("/element-event-workflows", elementEventWorkflowHandler.CreateElementEventWorkflow)
	group.Get("/element-event-workflows/:id", elementEventWorkflowHandler.GetElementEventWorkflowByID)
	group.Get("/element-event-workflows", elementEventWorkflowHandler.GetElementEventWorkflows)
	group.Patch("/element-event-workflows/:id", elementEventWorkflowHandler.UpdateElementEventWorkflow)
	group.Delete("/element-event-workflows/:id", elementEventWorkflowHandler.DeleteElementEventWorkflow)
	group.Delete("/element-event-workflows/element/:elementId", elementEventWorkflowHandler.DeleteElementEventWorkflowsByElement)
	group.Delete("/element-event-workflows/workflow/:workflowId", elementEventWorkflowHandler.DeleteElementEventWorkflowsByWorkflow)
}
