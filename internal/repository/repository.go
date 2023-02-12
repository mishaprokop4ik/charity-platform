package repository

type Repository struct {
	User          Userer
	Admin         adminCRUDer
	ProposalEvent ProposalEventer
}

func New(dbConnector *Connector) *Repository {
	return &Repository{
		User:          NewUser(dbConnector),
		Admin:         NewAdmin(dbConnector),
		ProposalEvent: NewProposalEvent(dbConnector),
	}
}
