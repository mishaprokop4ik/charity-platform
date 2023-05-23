package repository

import (
	"Kurajj/internal/models"
	"context"
	"io"
	"time"
)

//go:generate mockgen -source=repository.go -destination=mocks/mock.go

type HelpEventer interface {
	CreateEvent(ctx context.Context, event *models.HelpEvent) (uint, error)
	GetEventByID(ctx context.Context, id models.ID) (models.HelpEvent, error)
	GetHelpEventNeeds(ctx context.Context, eventID models.ID) ([]models.Need, error)
	CreateNeed(ctx context.Context, need models.Need) (uint, error)
	GetAllHelpEvents(ctx context.Context) ([]models.HelpEvent, error)
	UpdateNeeds(ctx context.Context, needs ...models.Need) error
	GetHelpEventByTransactionID(ctx context.Context, transactionID models.ID) (models.HelpEvent, error)
	UpdateHelpEvent(ctx context.Context, event models.HelpEvent) error
	GetUserHelpEvents(ctx context.Context, userID models.ID) ([]models.HelpEvent, error)
	GetHelpEventsWithSearchAndSort(ctx context.Context,
		searchValues models.HelpSearchInternal) (models.HelpEventPagination, error)
	GetTransactionNeeds(ctx context.Context, transactionID models.ID) ([]models.Need, error)
	GetHelpEventStatistics(ctx context.Context, id uint, from, to time.Time) ([]models.Transaction, error)
}

type AdminCRUDer interface {
	CreateAdmin(ctx context.Context, admin models.User) (uint, error)
	GetAdminByID(ctx context.Context, id uint) (models.User, error)
	UpdateAdmin(ctx context.Context, admin models.User) error
	DeleteAdmin(ctx context.Context, id uint) error
	GetAllAdmins(ctx context.Context) ([]models.User, error)
}

type Filer interface {
	Get(ctx context.Context, identifier string) (io.Reader, error)
	Upload(ctx context.Context, fileName string, fileData io.Reader) (string, error)
	Delete(ctx context.Context, identifier string) error
}

type Commenter interface {
	GetAllCommentsInEvent(ctx context.Context, eventID uint, eventType models.EventType) ([]models.Comment, error)
	GetCommentByID(ctx context.Context, id uint) (models.Comment, error)
	UpdateComment(ctx context.Context, id uint, toUpdate map[string]any) error
	DeleteComment(ctx context.Context, id uint) error
	WriteComment(ctx context.Context, comment models.Comment) (uint, error)
}

type ProposalEventer interface {
	CreateProposalEvent(ctx context.Context, event models.ProposalEvent) (uint, error)
	GetEvent(ctx context.Context, id uint) (models.ProposalEvent, error)
	GetEvents(ctx context.Context) ([]models.ProposalEvent, error)
	UpdateEvent(ctx context.Context, event models.ProposalEvent) error
	UpdateRemainingHelps(ctx context.Context, eventID models.ID, increase bool, number int) error
	DeleteEvent(ctx context.Context, id uint) error
	GetUserProposalEvents(ctx context.Context, userID uint) ([]models.ProposalEvent, error)
	GetProposalEventsWithSearchAndSort(ctx context.Context,
		searchValues models.ProposalEventSearchInternal) (models.ProposalEventPagination, error)
	GetProposalEventByTransactionID(ctx context.Context, transactionID int) (models.ProposalEvent, error)
	GetProposalEventStatistics(ctx context.Context, id uint, from, to time.Time) ([]models.Transaction, error)
}

type Tagger interface {
	UpsertTags(ctx context.Context, eventType models.EventType, eventID uint, tags []models.Tag) error
	GetTagsByEvent(ctx context.Context, eventID uint, eventType models.EventType) ([]models.Tag, error)
	DeleteAllTagsByEvent(ctx context.Context, eventID uint, eventType models.EventType) error
	CreateTag(ctx context.Context, tag models.Tag) error
}

type Transactioner interface {
	UpdateTransactionByEvent(ctx context.Context, eventID uint, eventType models.EventType, toUpdate map[string]any) error
	UpdateTransactionByID(ctx context.Context, id uint, toUpdate map[string]any) error
	GetCurrentEventTransactions(ctx context.Context,
		eventID uint,
		eventType models.EventType) ([]models.Transaction, error)
	UpdateAllNotFinishedTransactions(ctx context.Context, eventID uint, eventType models.EventType, newStatus models.TransactionStatus) error
	GetAllEventTransactions(ctx context.Context, eventID uint, eventType models.EventType) ([]models.Transaction, error)
	CreateTransaction(ctx context.Context, transaction models.Transaction) (uint, error)
	GetTransactionByID(ctx context.Context, id uint) (models.Transaction, error)
	GetGlobalStatistics(ctx context.Context, from, to time.Time) ([]models.Transaction, error)
}

type Notifier interface {
	CreateNotification(ctx context.Context, notification models.TransactionNotification) (uint, error)
	Update(ctx context.Context, newNotification models.TransactionNotification) error
	ReadNotifications(ctx context.Context, ids []uint) error
	GetByMember(ctx context.Context, userID uint) ([]models.TransactionNotification, error)
	GetByID(ctx context.Context, id uint) (models.TransactionNotification, error)
}

type Userer interface {
	CreateUser(ctx context.Context, user models.User) (uint, error)
	GetUserAuthentication(ctx context.Context, email, password string) (models.User, error)
	GetUserInfo(ctx context.Context, id uint) (models.User, error)
	GetEntity(ctx context.Context, email, password string, isAdmin, isDeleted bool) (models.User, error)
	SetSession(ctx context.Context, userID uint, session models.MemberSession) error
	GetByRefreshToken(ctx context.Context, token string) (models.User, error)
	DeleteUser(ctx context.Context, id uint) error
	UpsertUser(ctx context.Context, values map[string]any) error
	UpdateUserByEmail(ctx context.Context, email string, values map[string]any) error
	IsEmailTaken(ctx context.Context, email string) (bool, error)
}

type UserSearcher interface {
	UpsertUserTags(ctx context.Context, userID uint, searchValues []models.MemberSearch) error
}

type Complainer interface {
	Complain(ctx context.Context, complaint models.Complaint) (int, error)
	GetAll(ctx context.Context) ([]models.ComplaintsResponse, error)
	BanUser(ctx context.Context, userID models.ID) error
	BanEvent(ctx context.Context, eventID models.ID, eventType models.EventType) error
}

type Repository struct {
	Userer
	AdminCRUDer
	ProposalEventer
	Transactioner
	Commenter
	Tagger
	UserSearcher
	Filer
	Notifier
	HelpEventer
	Complainer
}

func New(dbConnector *Connector, config AWSConfig) *Repository {
	return &Repository{
		NewUser(config, dbConnector),
		NewAdmin(dbConnector),
		NewProposalEvent(config, dbConnector),
		NewTransaction(dbConnector),
		NewComment(dbConnector),
		NewTag(dbConnector),
		NewUserSearch(dbConnector),
		NewFile(config),
		NewTransactionNotification(dbConnector),
		NewHelpEvent(config, dbConnector),
		NewComplaint(dbConnector),
	}
}
