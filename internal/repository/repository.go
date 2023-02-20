package repository

type Repository struct {
	User          Userer
	Admin         adminCRUDer
	ProposalEvent ProposalEventer
	Transaction   Transactioner
}

func New(dbConnector *Connector) *Repository {
	return &Repository{
		User:          NewUser(dbConnector),
		Admin:         NewAdmin(dbConnector),
		ProposalEvent: NewProposalEvent(dbConnector),
		Transaction:   NewTransaction(dbConnector),
	}
}
