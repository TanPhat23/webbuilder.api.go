package testutil

import (
	"my-go-app/internal/repositories"
)

var (
	_ repositories.UserRepositoryInterface                 = (*MockUserRepository)(nil)
	_ repositories.ProjectRepositoryInterface              = (*MockProjectRepository)(nil)
	_ repositories.ImageRepositoryInterface                = (*MockImageRepository)(nil)
	_ repositories.PageRepositoryInterface                 = (*MockPageRepository)(nil)
	_ repositories.SnapshotRepositoryInterface             = (*MockSnapshotRepository)(nil)
	_ repositories.CollaboratorRepositoryInterface         = (*MockCollaboratorRepository)(nil)
	_ repositories.EventWorkflowRepositoryInterface        = (*MockEventWorkflowRepository)(nil)
	_ repositories.ElementEventWorkflowRepositoryInterface = (*MockElementEventWorkflowRepository)(nil)
	_ repositories.ElementRepositoryInterface              = (*MockElementRepository)(nil)
	_ repositories.CustomElementRepositoryInterface        = (*MockCustomElementRepository)(nil)
	_ repositories.CustomElementTypeRepositoryInterface    = (*MockCustomElementTypeRepository)(nil)
	_ repositories.ElementCommentRepositoryInterface       = (*MockElementCommentRepository)(nil)
	_ repositories.ContentTypeRepositoryInterface          = (*MockContentTypeRepository)(nil)
	_ repositories.ContentFieldRepositoryInterface         = (*MockContentFieldRepository)(nil)
	_ repositories.ContentItemRepositoryInterface          = (*MockContentItemRepository)(nil)
	_ repositories.ContentFieldValueRepositoryInterface    = (*MockContentFieldValueRepository)(nil)
	_ repositories.MarketplaceRepositoryInterface          = (*MockMarketplaceRepository)(nil)

	_ repositories.CommentRepositoryInterface              = (*MockCommentRepository)(nil)
	_ repositories.InvitationRepositoryInterface           = (*MockInvitationRepository)(nil)
)