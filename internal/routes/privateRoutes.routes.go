package routes

import (
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"
	"my-go-app/pkg/middleware"

	// "my-go-app/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

func PrivateRoutes(app *fiber.App, repos *repositories.RepositoriesInterface) {
	elementHandler := handlers.NewElementHandler(repos.ElementRepository)
	projectHandler := handlers.NewProjectHandler(repos.ProjectRepository)
	pageHandler := handlers.NewPageHandler(repos.PageRepository)
	snapshotHandler := handlers.NewSnapshotHandler(repos.SnapshotRepository, repos.ElementRepository)

	group := app.Group("/api/v1", middleware.AuthenticateMiddleware)

	group.Get("/elements/:projectid", elementHandler.GetElements)
	group.Post("/elements/:projectid", elementHandler.CreateElements)
	group.Post("/elements/:projectid/insert/:previouselementid", elementHandler.InsertElementAfter)
	group.Patch("/elements/:elementid", elementHandler.UpdateElement)
	group.Delete("/elements/:elementid", elementHandler.DeleteElement)

	group.Get("/projects/user", projectHandler.GetProjectByUserID)
	group.Get("/projects/:projectid", projectHandler.GetProjectByID)
	group.Get("/projects/:projectid/pages", projectHandler.GetProjectPages)
	group.Delete("/projects/:projectid", projectHandler.DeleteProject)
	group.Patch("/projects/:projectid", projectHandler.UpdateProject)
	group.Delete("/projects/:projectid/pages/:pageid", pageHandler.DeletePage)

	group.Post("/snapshots/:projectid/save", snapshotHandler.SaveSnapshot)

}
