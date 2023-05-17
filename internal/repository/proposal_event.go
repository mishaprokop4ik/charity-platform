package repository

import (
	"Kurajj/internal/models"
	zlog "Kurajj/pkg/logger"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"gorm.io/gorm"
	"math"
	"strings"
	"time"
)

const defaultSortField = "creation_date"

const defaultProposalImage = "https://charity-platform.s3.amazonaws.com/images/volunteer-care-old-people-nurse-isolated-young-human-helping-senior-volunteers-service-helpful-person-nursing-elderly-decent-vector-set_53562-17770.avif"

type ProposalEvent struct {
	DBConnector *Connector
	Filer
}

func (p *ProposalEvent) UpdateRemainingHelps(ctx context.Context, eventID models.ID, increase bool, number int) error {
	oldEvent, err := p.GetEvent(ctx, uint(eventID))
	if err != nil {
		return err
	}
	if increase {
		oldEvent.RemainingHelps += number
	} else {
		oldEvent.RemainingHelps -= number
	}
	err = p.DBConnector.DB.
		Where("id = ?", eventID).
		Model(&models.ProposalEvent{}).
		Select("RemainingHelps").
		Updates(&oldEvent).
		WithContext(ctx).
		Error
	return err
}

func (p *ProposalEvent) GetProposalEventStatistics(ctx context.Context, creatorID uint, from, to time.Time) ([]models.Transaction, error) {
	proposalEvents := []models.ProposalEvent{}
	err := p.DBConnector.DB.
		Where("author_id = ?", creatorID).
		Find(&proposalEvents).
		WithContext(ctx).Error
	if err != nil {
		return nil, err
	}

	proposalEventIds := make([]uint, len(proposalEvents))
	for i := range proposalEvents {
		proposalEventIds[i] = proposalEvents[i].ID
	}
	transactions := []models.Transaction{}
	err = p.DBConnector.DB.
		Where("event_type = ?", models.ProposalEventType).
		Where("event_id IN (?)", proposalEventIds).
		Where("creation_date >= ? AND creation_date <= ?",
			from, to).
		Find(&transactions).
		WithContext(ctx).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		zlog.Log.Error(err, "could not get any transactions")
		return transactions, nil
	}

	return transactions, err
}

func (p *ProposalEvent) GetProposalEventByTransactionID(ctx context.Context, transactionID int) (models.ProposalEvent, error) {
	transaction := models.Transaction{}
	err := p.DBConnector.DB.Where("id = ?", transactionID).First(&transaction).WithContext(ctx).Error
	if err != nil {
		return models.ProposalEvent{}, err
	}
	event := models.ProposalEvent{}
	err = p.DBConnector.DB.Where("id = ?", transaction.EventID).First(&event).WithContext(ctx).Error
	return event, err
}

const (
	DefaultPageNumber = 1
	DefaultPageLimit  = 10
)

func (p *ProposalEvent) calculatePagination(ctx context.Context, searchValues models.ProposalEventSearchInternal, searchQuery *gorm.DB) (*models.Pagination, error) {
	offset := 0

	if searchValues.Pagination.PageNumber > 0 {
		offset = (searchValues.Pagination.PageNumber - 1) * searchValues.Pagination.PageSize
	}

	events := []models.ProposalEvent{}
	err := searchQuery.Find(&events).Distinct().WithContext(ctx).Error
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(len(events)) / float64(searchValues.Pagination.PageSize)))

	pagination := &models.Pagination{
		TotalRecords: int64(len(events)),
		TotalPage:    totalPages,
		Offset:       offset,
		Limit:        searchValues.Pagination.PageSize,
		Page:         searchValues.Pagination.PageNumber,
		PrevPage:     searchValues.Pagination.PageNumber,
		NextPage:     searchValues.Pagination.PageNumber,
	}

	if searchValues.Pagination.PageNumber > 1 {
		pagination.PrevPage = searchValues.Pagination.PageNumber - 1
	}

	if searchValues.Pagination.PageNumber != pagination.TotalPage {
		pagination.NextPage = searchValues.Pagination.PageNumber + 1
	}

	return pagination, nil
}

func (p *ProposalEvent) GetProposalEventsWithSearchAndSort(ctx context.Context,
	searchValues models.ProposalEventSearchInternal) (models.ProposalEventPagination, error) {
	db := p.DBConnector.DB.Session(&gorm.Session{})
	events := []models.ProposalEvent{}
	searchValues = p.removeEmptySearchValues(searchValues)
	query := db.
		Order(fmt.Sprintf("propositional_event.%s %s", searchValues.SortField, strings.ToUpper(string(*searchValues.Order)))).
		Where("status IN (?)", searchValues.State)
	query = query.Debug()

	if searchValues.Name != nil && *searchValues.Name != "" {
		query = query.Where("LOWER(title) LIKE ?", "%"+*searchValues.Name+"%")
	}
	if searchValues.SearcherID != nil {
		query = query.Not("propositional_event.author_id = ?", searchValues.SearcherID)
	}

	if searchValues.TakingPart != nil && searchValues.SearcherID != nil {
		if *searchValues.TakingPart {
			query = query.Joins("LEFT JOIN transaction ON propositional_event.id = transaction.event_id").
				Where("transaction.creator_id = ?", searchValues.SearcherID).
				Distinct()
		}
	}

	if searchValues.Tags != nil && len(*searchValues.Tags) != 0 {
		if *searchValues.AllowTitleSearch && len(searchValues.GetTagsValues()) == 0 && len(searchValues.GetTagsTitle()) != 0 {
			subQuery := db.Table("tag").Select("event_id").
				Where("LOWER(title) IN (?) AND event_type = ?", searchValues.GetTagsTitle(), models.ProposalEventType)

			query = query.Where("propositional_event.id IN (?)", subQuery)
		} else if len(searchValues.GetTagsValues()) != 0 && len(searchValues.GetTagsTitle()) != 0 {
			subQuery := db.Table("tag").Select("event_id").
				Joins("JOIN tag_value ON tag.id = tag_value.tag_id").
				Where("LOWER(tag.title) IN (?) AND LOWER(tag_value.value) IN (?) AND tag.event_type = ?",
					searchValues.GetTagsTitle(), searchValues.GetTagsValues(), models.ProposalEventType)

			query = query.Where("propositional_event.id IN (?)", subQuery)
		}
	}
	if searchValues.Location != nil && searchValues.Location.String() != "|||" && searchValues.Location.Values() != "" {
		location := *searchValues.Location
		subQuery := db.Table("location").Select("event_id").
			Where("event_type = ?", models.ProposalEventType)

		if location.Region != "" {
			subQuery = subQuery.Where("LOWER(area) LIKE ?", "%"+location.Region+"%")
		}
		if location.City != "" {
			subQuery = subQuery.Where("LOWER(city) LIKE ?", "%"+location.City+"%")
		}
		if location.District != "" {
			subQuery = subQuery.Where("LOWER(district) LIKE ?", "%"+location.District+"%")
		}
		if location.HomeLocation != "" {
			subQuery = subQuery.Where("LOWER(home) LIKE ?", "%"+location.HomeLocation+"%")
		}

		query = query.Where("propositional_event.id IN (?)", subQuery)
	}

	pagination, err := p.calculatePagination(ctx, searchValues, query)
	if err != nil {
		zlog.Log.Error(err, "could not calculate pagination value")
		return models.ProposalEventPagination{}, err
	}

	err = query.Limit(searchValues.Pagination.PageSize).Offset(pagination.Offset).Find(&events).WithContext(ctx).Error
	if err != nil {
		return models.ProposalEventPagination{}, err
	}

	events, err = p.insertUserInProposalEvents(ctx, events)
	if err != nil {
		return models.ProposalEventPagination{}, err
	}

	for i, event := range events {
		updatedEvent, err := p.updateMissingEventData(ctx, event)
		if err != nil {
			return models.ProposalEventPagination{}, err
		}
		events[i] = updatedEvent
	}

	return models.ProposalEventPagination{
		Events:     events,
		Pagination: *pagination,
	}, nil
}

func (p *ProposalEvent) removeEmptySearchValues(searchValues models.ProposalEventSearchInternal) models.ProposalEventSearchInternal {
	boolRef := func(b bool) *bool {
		return &b
	}
	newSearchValues := models.ProposalEventSearchInternal{}
	if searchValues.SortField == "" {
		newSearchValues.SortField = defaultSortField
	} else {
		newSearchValues.SortField = searchValues.SortField
	}

	if searchValues.Name != nil {
		newSearchValues.Name = searchValues.Name
	}

	if searchValues.Tags != nil {
		newSearchValues.Tags = searchValues.Tags
	} else {
		newSearchValues.Tags = nil
	}

	newSearchValues.SearcherID = searchValues.SearcherID

	if searchValues.TakingPart != nil {
		newSearchValues.TakingPart = searchValues.TakingPart
	}

	if searchValues.State == nil {
		newSearchValues.State = []models.EventStatus{
			models.Active,
			models.Done,
		}
	} else {
		newSearchValues.State = searchValues.State
	}

	if searchValues.Location == nil {
		newSearchValues.Location = nil
	} else {
		newSearchValues.Location = searchValues.Location
	}
	if searchValues.Order == nil || *searchValues.Order == "" {
		newSearchValues.Order = &models.AscendingOrder
	} else {
		newSearchValues.Order = searchValues.Order
	}

	if searchValues.Pagination.PageNumber < 1 {
		newSearchValues.Pagination.PageNumber = DefaultPageNumber
	} else {
		newSearchValues.Pagination.PageNumber = searchValues.Pagination.PageNumber
	}

	if searchValues.Pagination.PageSize < 1 {
		newSearchValues.Pagination.PageSize = DefaultPageLimit
	} else {
		newSearchValues.Pagination.PageSize = searchValues.Pagination.PageSize
	}

	if searchValues.AllowTitleSearch == nil {
		newSearchValues.AllowTitleSearch = boolRef(false)
	} else {
		newSearchValues.AllowTitleSearch = searchValues.AllowTitleSearch
	}

	return newSearchValues
}

func (p *ProposalEvent) CreateProposalEvent(ctx context.Context, event models.ProposalEvent) (uint, error) {
	tx := p.DBConnector.DB.Begin()

	if event.File != nil {
		fileName, err := uuid.NewUUID()
		if err != nil {
			tx.Commit()
			return 0, err
		}
		filePath, err := p.Filer.Upload(ctx, fmt.Sprintf("%s.%s", fileName.String(), event.FileType), event.File)
		if err != nil {
			zlog.Log.Error(err, "could not upload file")
			return 0, err
		}
		event.ImagePath = filePath
	} else {
		event.ImagePath = defaultProposalImage
	}

	err := tx.
		Create(&event).
		WithContext(ctx).
		Error

	if err != nil {
		tx.Rollback()
		return 0, err
	}
	if !event.Location.IsEmpty() {
		event.Location.EventID = event.ID
		err = tx.
			Create(&event.Location).
			WithContext(ctx).
			Error

		if err != nil {
			tx.Rollback()
			return 0, err
		}
	}
	for _, tag := range event.Tags {
		tag.EventID = event.ID
		err = tx.Create(&tag).WithContext(ctx).Error
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		fmt.Println(tag, "123proposal123")
		for _, tagValue := range tag.Values {
			tagValue.TagID = tag.ID
			fmt.Println(tagValue, "321proposal321")
			err = tx.Create(&tagValue).WithContext(ctx).Error
			if err != nil {
				tx.Rollback()
				return 0, err
			}
		}
	}

	return event.ID, tx.Commit().Error
}

func (p *ProposalEvent) GetEvent(ctx context.Context, id uint) (models.ProposalEvent, error) {
	event := models.ProposalEvent{}
	err := p.DBConnector.DB.
		Where("id = ?", id).
		Where("is_deleted = ?", false).
		First(&event).
		WithContext(ctx).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.ProposalEvent{}, fmt.Errorf("cound not get proposal event by id %d", id)
	} else if err != nil {
		return models.ProposalEvent{}, err
	}

	event, err = p.updateMissingEventData(ctx, event)
	if err != nil {
		return models.ProposalEvent{}, nil
	}

	events, err := p.insertUserInProposalEvents(ctx, []models.ProposalEvent{
		event,
	})
	if err != nil {
		return models.ProposalEvent{}, err
	}
	return events[0], nil
}

func (p *ProposalEvent) GetEvents(ctx context.Context) ([]models.ProposalEvent, error) {
	events := []models.ProposalEvent{}
	resp := p.DBConnector.DB.
		Where("is_deleted = ?", false).
		Find(&events).
		WithContext(ctx)

	if errors.Is(resp.Error, gorm.ErrRecordNotFound) {
		return []models.ProposalEvent{}, fmt.Errorf("could not get any proposal events")
	} else if resp.Error != nil {
		return nil, resp.Error
	}

	events, err := p.insertUserInProposalEvents(ctx, events)
	if err != nil {
		return nil, err
	}

	for i, event := range events {
		newEvent, err := p.updateMissingEventData(ctx, event)
		if err != nil {
			return nil, err
		}
		events[i] = newEvent
	}

	return events, nil
}

func (p *ProposalEvent) UpdateEvent(ctx context.Context, event models.ProposalEvent) error {
	if event.File != nil {
		if err := p.saveFile(ctx, &event); err != nil {
			return err
		}
	}
	fmt.Println(event.GetValuesToUpdate(), lo.Keys(event.GetValuesToUpdate()))
	return p.DBConnector.DB.
		Model(&models.ProposalEvent{}).
		Select(lo.Keys(event.GetValuesToUpdate())).
		Where("id = ?", event.ID).
		Updates(event.GetValuesToUpdate()).
		WithContext(ctx).
		Error
}
func (p *ProposalEvent) UpdateEvent2(ctx context.Context, event models.ProposalEvent) error {
	if event.File != nil {
		if err := p.saveFile(ctx, &event); err != nil {
			return err
		}
	}
	err := p.DBConnector.DB.Model(&event).Updates(event).Where("id = ?", event.ID).WithContext(ctx).Error
	return err
}

func (p *ProposalEvent) saveFile(ctx context.Context, event *models.ProposalEvent) error {
	if event.File != nil {
		oldEvent := models.ProposalEvent{}
		err := p.DBConnector.DB.Where("id = ?", event.ID).First(&oldEvent).WithContext(ctx).Error
		if err != nil {
			return err
		}

		imagePath := strings.Split(oldEvent.ImagePath, "/")
		imageName := imagePath[len(imagePath)-1]
		err = p.Filer.Delete(ctx, imageName)
		if err != nil {
			return err
		}
		fileName, err := uuid.NewUUID()
		if err != nil {
			return err
		}
		filePath, err := p.Filer.Upload(ctx, fmt.Sprintf("%s.%s", fileName.String(), event.FileType), event.File)
		if err != nil {
			zlog.Log.Error(err, "could not upload file")
			return err
		}
		event.ImagePath = filePath
	}
	return nil
}

func (p *ProposalEvent) DeleteEvent(ctx context.Context, id uint) error {
	tx := p.DBConnector.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	oldProposalEvent := &models.ProposalEvent{}
	err := tx.Where("id = ?", id).WithContext(ctx).First(oldProposalEvent).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	oldProposalEvent.CompetitionDate = sql.NullTime{Time: time.Now(), Valid: true}
	oldProposalEvent.IsDeleted = true
	err = tx.Where("id = ?", id).Updates(oldProposalEvent).WithContext(ctx).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.
		Model(&models.Transaction{}).
		Where("event_id = ?", id).
		Where("event_type = ?", models.ProposalEventType).
		Not("status IN (?)", []models.TransactionStatus{models.Completed, models.Interrupted, models.Canceled}).
		Update("status", models.Canceled).
		WithContext(ctx).
		Error
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.
		Model(&models.Comment{}).
		Where("event_id = ?", id).
		Where("event_type = ?", models.ProposalEventType).
		Update("is_deleted", true).
		WithContext(ctx).
		Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (p *ProposalEvent) GetUserProposalEvents(ctx context.Context, userID uint) ([]models.ProposalEvent, error) {
	events := []models.ProposalEvent{}
	resp := p.DBConnector.DB.
		Where("author_id = ?", userID).
		Where("is_deleted", false).
		Find(&events).
		WithContext(ctx)

	if errors.Is(resp.Error, gorm.ErrRecordNotFound) {
		return []models.ProposalEvent{}, fmt.Errorf("could not get any user proposal events")
	} else if resp.Error != nil {
		return nil, resp.Error
	}

	events, err := p.insertUserInProposalEvents(ctx, events)
	if err != nil {
		return nil, err
	}

	for i, event := range events {
		newEvent, err := p.updateMissingEventData(ctx, event)
		if err != nil {
			return nil, err
		}
		events[i] = newEvent
	}

	return events, nil
}

func (p *ProposalEvent) insertUserInProposalEvents(ctx context.Context, events []models.ProposalEvent) ([]models.ProposalEvent, error) {
	for i, event := range events {
		memberID := event.AuthorID
		member := models.User{}
		err := p.DBConnector.DB.Where("id = ?", memberID).First(&member).WithContext(ctx).Error
		if err != nil {
			return []models.ProposalEvent{}, err
		}
		events[i].User = member
	}
	return events, nil
}

func (p *ProposalEvent) getProposalEventTransactions(ctx context.Context, eventID uint) ([]models.Transaction, error) {
	transactions := []models.Transaction{}
	err := p.DBConnector.DB.
		Where("event_id = ?", eventID).
		Where("event_type = ?", models.ProposalEventType).
		Find(&transactions).
		WithContext(ctx).
		Error
	if err != nil {
		return []models.Transaction{}, err
	}
	for i, transaction := range transactions {
		newTransaction, err := p.updateTransactionUsers(ctx, transaction)
		if err != nil {
			return nil, err
		}
		transactions[i] = newTransaction
	}
	return transactions, err
}

func (p *ProposalEvent) updateTransactionUsers(ctx context.Context, transaction models.Transaction) (models.Transaction, error) {
	creatorInfo := models.User{}
	err := p.DBConnector.DB.Where("id = ?", transaction.CreatorID).First(&creatorInfo).WithContext(ctx).Error
	if err != nil {
		return models.Transaction{}, err
	}

	transaction.Creator = creatorInfo

	rootEvent := models.ProposalEvent{}

	err = p.DBConnector.DB.Where("id = ?", transaction.EventID).First(&rootEvent).WithContext(ctx).Error
	if err != nil {
		return models.Transaction{}, err
	}

	responderInfo := models.User{}
	err = p.DBConnector.DB.Where("id = ?", rootEvent.AuthorID).First(&responderInfo).WithContext(ctx).Error
	if err != nil {
		return models.Transaction{}, err
	}

	transaction.Responder = responderInfo

	return transaction, nil
}

func (p *ProposalEvent) getProposalEventComments(ctx context.Context, eventID uint) ([]models.Comment, error) {
	comments := []models.Comment{}
	err := p.DBConnector.DB.
		Where("event_type = ?", models.ProposalEventType).
		Where("event_id = ?", eventID).
		Where("is_deleted = ?", false).
		Find(&comments).
		WithContext(ctx).
		Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return comments, err
	}

	for i, comment := range comments {
		newComment, err := p.updateCommentUsers(ctx, comment)
		if err != nil {
			return nil, err
		}
		comments[i] = newComment
	}

	return comments, nil
}

func (p *ProposalEvent) updateCommentUsers(ctx context.Context, comment models.Comment) (models.Comment, error) {
	creatorInfo := models.User{}
	err := p.DBConnector.DB.Where("id = ?", comment.UserID).First(&creatorInfo).WithContext(ctx).Error
	if err != nil {
		return models.Comment{}, err
	}

	comment.UserValues = models.UserShortInfo{
		ID:              creatorInfo.ID,
		Username:        creatorInfo.FullName,
		ProfileImageURL: creatorInfo.AvatarImagePath,
		PhoneNumber:     models.Telephone(creatorInfo.Telephone),
	}

	return comment, nil
}

func (p *ProposalEvent) updateMissingEventData(ctx context.Context, proposalEvent models.ProposalEvent) (models.ProposalEvent, error) {
	comments, err := p.getProposalEventComments(ctx, proposalEvent.ID)
	if err != nil {
		return models.ProposalEvent{}, err
	}
	proposalEvent.Comments = comments
	transactions, err := p.getProposalEventTransactions(ctx, proposalEvent.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.ProposalEvent{}, err
	}
	proposalEvent.Transactions = transactions
	location := models.Address{}
	err = p.DBConnector.DB.
		Where("event_type = ?", models.ProposalEventType).
		Where("event_id = ?", proposalEvent.ID).
		Find(&location).
		WithContext(ctx).
		Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.ProposalEvent{}, err
	} else if err == nil {
		proposalEvent.Location = location
	}

	tags := []models.Tag{}
	err = p.DBConnector.DB.
		Where("event_type = ?", models.ProposalEventType).
		Where("event_id = ?", proposalEvent.ID).
		Find(&tags).
		WithContext(ctx).
		Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.ProposalEvent{}, err
	}
	proposalEvent.Tags = tags

	for i, tag := range proposalEvent.Tags {
		tagValues := []models.TagValue{}
		err = p.DBConnector.DB.Where("tag_id = ?", tag.ID).
			Find(&tagValues).
			WithContext(ctx).
			Error

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			zlog.Log.Error(err, "got error while getting tag values")
			continue
		}
		proposalEvent.Tags[i].Values = tagValues
		tag.Values = tagValues
	}

	return proposalEvent, nil
}

func NewProposalEvent(config AWSConfig, DBConnector *Connector) *ProposalEvent {
	return &ProposalEvent{DBConnector: DBConnector, Filer: NewFile(config)}
}
