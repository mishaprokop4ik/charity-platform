package repository

type Repository struct {
	User            Userer
	Admin           adminCRUDer
	ProposalEvent   ProposalEventer
	Transaction     Transactioner
	Comment         Commenter
	Tag             Tagger
	UserSearchValue UserSearcher
}

func New(dbConnector *Connector, config AWSConfig) *Repository {
	return &Repository{
		User:            NewUser(dbConnector),
		Admin:           NewAdmin(dbConnector),
		ProposalEvent:   NewProposalEvent(config, dbConnector),
		Transaction:     NewTransaction(dbConnector),
		Comment:         NewComment(dbConnector),
		Tag:             NewTag(dbConnector),
		UserSearchValue: NewUserSearch(dbConnector),
	}
}
