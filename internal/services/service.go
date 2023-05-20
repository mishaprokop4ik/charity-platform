package service

import (
	"Kurajj/configs"
	"Kurajj/internal/models"
	"Kurajj/internal/repository"
	"context"
	"io"
)

//go:generate mockgen -source=service.go -destination=mocks/service.go

type Repositorier interface {
	repository.Userer
	repository.AdminCRUDer
	repository.ProposalEventer
	repository.Transactioner
	repository.Commenter
	repository.Tagger
	repository.UserSearcher
	repository.Filer
	repository.Notifier
	repository.HelpEventer
	repository.Complainer
}

type HelpEventer interface {
	CreateHelpEvent(ctx context.Context, event *models.HelpEvent) (uint, error)
	GetHelpEventByID(ctx context.Context, id models.ID) (models.HelpEvent, error)
	GetHelpEventByTransactionID(ctx context.Context, transactionID models.ID) (models.HelpEvent, error)
	CreateRequest(ctx context.Context, userID models.ID, transactionInfo models.TransactionAcceptCreateRequest) (uint, error)
	UpdateTransactionStatus(ctx context.Context, transaction models.HelpEventTransaction, file io.Reader, fileType, createFilePath string) error
	GetUserHelpEvents(ctx context.Context, userID models.ID) ([]models.HelpEvent, error)
	GetHelpEventBySearch(ctx context.Context, search models.HelpSearchInternal) (models.HelpEventPagination, error)
	UpdateHelpEvent(ctx context.Context, event models.HelpEvent) error
	GetHelpEventStatistics(ctx context.Context, fromStart int, creatorID uint) (models.HelpEventStatistics, error)
}

type ProposalEventer interface {
	CreateEvent(ctx context.Context, event models.ProposalEvent) (uint, error)
	GetEvent(ctx context.Context, id uint) (models.ProposalEvent, error)
	GetEvents(ctx context.Context) ([]models.ProposalEvent, error)
	UpdateProposalEvent(ctx context.Context, event models.ProposalEvent) error
	DeleteEvent(ctx context.Context, id uint) error
	Response(ctx context.Context, proposalEventID, responderID uint, comment string) error
	Accept(ctx context.Context, request models.AcceptRequest) error
	UpdateStatus(ctx context.Context, status models.TransactionStatus, transactionID, userID uint, file io.Reader, fileType, filePath string) error
	GetUserProposalEvents(ctx context.Context, userID uint) ([]models.ProposalEvent, error)
	GetProposalEventBySearch(ctx context.Context, search models.ProposalEventSearchInternal) (models.ProposalEventPagination, error)
	GetProposalEventStatistics(ctx context.Context, fromStart int, creatorID uint) (models.ProposalEventStatistics, error)
}

type AdminCRUDer interface {
	CreateAdmin(ctx context.Context, admin models.User) (uint, error)
	GetAdminByID(ctx context.Context, id uint) (models.User, error)
	UpdateAdmin(ctx context.Context, admin models.User) error
	DeleteAdmin(ctx context.Context, id uint) error
	GetAllAdmins(ctx context.Context) ([]models.User, error)
}

type Transactioner interface {
	UpdateTransaction(ctx context.Context, transaction models.Transaction) error
	GetCurrentEventTransactions(ctx context.Context,
		eventID uint,
		eventType models.EventType) ([]models.Transaction, error)
	UpdateAllNotFinishedTransactions(ctx context.Context, eventID uint, eventType models.EventType, newStatus models.TransactionStatus) error
	GetAllEventTransactions(ctx context.Context, eventID uint, eventType models.EventType) ([]models.Transaction, error)
	CreateTransaction(ctx context.Context, transaction models.Transaction) (uint, error)
	GetTransactionByID(ctx context.Context, id uint) (models.Transaction, error)
}

type Commenter interface {
	GetAllCommentsInEvent(ctx context.Context, eventID uint, eventType models.EventType) ([]models.Comment, error)
	GetCommentByID(ctx context.Context, id uint) (models.Comment, error)
	UpdateComment(ctx context.Context, comment models.Comment) error
	DeleteComment(ctx context.Context, id uint) error
	WriteComment(ctx context.Context, comment models.Comment) (uint, error)
}

type Tagger interface {
	UpsertTags(ctx context.Context, eventID uint, eventType models.EventType, tags []models.Tag) error
	GetTagsByEvent(ctx context.Context, eventID uint, eventType models.EventType) ([]models.Tag, error)
}

type UserSearcher interface {
	UpsertValues(ctx context.Context, userId uint, tags []models.MemberSearch) error
}

type TransactionNotifier interface {
	Read(ctx context.Context, id []uint) error
	GetUserNotifications(ctx context.Context, userID uint) ([]models.TransactionNotification, error)
}

type Filer interface {
	Get(ctx context.Context, identifier string) (io.Reader, error)
	Upload(ctx context.Context, fileName string, fileData io.Reader) (string, error)
	Delete(ctx context.Context, identifier string) error
}

type Complainer interface {
	Complain(ctx context.Context, complaint models.Complaint) (int, error)
	GetAll(ctx context.Context) ([]models.ComplaintsResponse, error)
	BanUser(ctx context.Context, userID models.ID) error
	BanEvent(ctx context.Context, eventID models.ID, eventType models.EventType) error
}

type Service struct {
	Authenticator
	AdminCRUDer
	ProposalEventer
	Transactioner
	Commenter
	Tagger
	UserSearcher
	TransactionNotifier
	HelpEventer
	Filer
	Complainer
}

func New(repo Repositorier,
	authConfig *configs.AuthenticationConfig,
	emailConfig *configs.Email,
) *Service {
	return &Service{
		NewAuthentication(repo, authConfig, emailConfig),
		NewAdmin(repo, authConfig, emailConfig),
		NewProposalEvent(repo),
		NewTransaction(repo),
		NewComment(repo),
		NewTag(repo),
		NewUserSearch(repo),
		NewTransactionNotification(repo),
		NewHelpEvent(repo),
		NewFile(repo),
		NewComplaint(repo),
	}
}
