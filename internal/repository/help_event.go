package repository

import (
	"Kurajj/internal/models"
	zlog "Kurajj/pkg/logger"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"math"
	"strings"
	"time"
)

func NewHelpEvent(config AWSConfig, DBConnector *Connector) *HelpEvent {
	return &HelpEvent{Connector: DBConnector, Filer: NewFile(config)}
}

type HelpEvent struct {
	*Connector
	Filer
}

func (h *HelpEvent) GetHelpEventStatistics(ctx context.Context, creatorID uint, from, to time.Time) ([]models.Transaction, error) {
	transactions := []models.Transaction{}
	err := h.DB.
		Where("event_type = ?", models.HelpEventType).
		Where("creator_id IN (?)", creatorID).
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

func (h *HelpEvent) GetTransactionNeeds(ctx context.Context, transactionID models.ID) ([]models.Need, error) {
	needs := make([]models.Need, 0)
	err := h.DB.Where("transaction_id = ?", transactionID).Find(&needs).WithContext(ctx).Error
	return needs, err
}

func (h *HelpEvent) GetHelpEventsWithSearchAndSort(ctx context.Context,
	searchValues models.HelpSearchInternal) (models.HelpEventPagination, error) {
	db := h.DB.Session(&gorm.Session{})
	events := make([]models.HelpEvent, 0)
	searchValues = h.removeEmptySearchValues(searchValues)
	query := db.
		Order(fmt.Sprintf("help_event.%s %s", searchValues.SortField, strings.ToUpper(string(*searchValues.Order)))).
		Where("status IN (?)", searchValues.State)
	query = query.Debug()

	if searchValues.Name != nil && *searchValues.Name != "" {
		query = query.Where("LOWER(title) LIKE ?", "%"+*searchValues.Name+"%")
	}
	if searchValues.SearcherID != nil {
		query = query.Not("help_event.created_by = ?", searchValues.SearcherID)
	}

	if searchValues.TakingPart != nil && searchValues.SearcherID != nil {
		if *searchValues.TakingPart {
			query = query.Joins("LEFT JOIN transaction ON help_event.id = transaction.event_id").
				Where("transaction.creator_id = ?", searchValues.SearcherID).
				Distinct()
		}
	}

	if searchValues.Tags != nil && len(*searchValues.Tags) != 0 {
		if *searchValues.AllowTitleSearch && len(searchValues.GetTagsValues()) == 0 && len(searchValues.GetTagsTitle()) != 0 {
			subQuery := db.Table("tag").Select("event_id").
				Where("LOWER(title) IN (?) AND event_type = ?", searchValues.GetTagsTitle(), models.HelpEventType)

			query = query.Where("help_event.id IN (?)", subQuery)
		} else if len(searchValues.GetTagsValues()) != 0 && len(searchValues.GetTagsTitle()) != 0 {
			subQuery := db.Table("tag").Select("event_id").
				Joins("JOIN tag_value ON tag.id = tag_value.tag_id").
				Where("LOWER(tag.title) IN (?) AND LOWER(tag_value.value) IN (?) AND tag.event_type = ?",
					searchValues.GetTagsTitle(), searchValues.GetTagsValues(), models.HelpEventType)

			query = query.Where("help_event.id IN (?)", subQuery)
		}
	}
	if searchValues.Location != nil && searchValues.Location.String() != "|||" && searchValues.Location.Values() != "" {
		location := *searchValues.Location
		subQuery := db.Table("location").Select("event_id").
			Where("event_type = ?", models.HelpEventType)

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

		query = query.Where("help_event.id IN (?)", subQuery)
	}

	pagination, err := h.calculatePagination(ctx, searchValues, query)
	if err != nil {
		zlog.Log.Error(err, "could not calculate pagination value")
		return models.HelpEventPagination{}, err
	}

	err = query.Limit(searchValues.Pagination.PageSize).Offset(pagination.Offset).Find(&events).WithContext(ctx).Error
	if err != nil {
		return models.HelpEventPagination{}, err
	}

	for i := range events {
		err := h.insertHelpEventMissingData(ctx, &events[i])
		if err != nil {
			return models.HelpEventPagination{}, err
		}
	}

	return models.HelpEventPagination{
		Events:     events,
		Pagination: *pagination,
	}, nil
}

func (h *HelpEvent) removeEmptySearchValues(searchValues models.HelpSearchInternal) models.HelpSearchInternal {
	boolRef := func(b bool) *bool {
		return &b
	}
	newSearchValues := models.HelpSearchInternal{}
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

func (h *HelpEvent) GetUserHelpEvents(ctx context.Context, userID models.ID) ([]models.HelpEvent, error) {
	helpEvents := make([]models.HelpEvent, 0)
	err := h.DB.Where("is_deleted = ?", false).Where("is_banned = ?", false).Where("created_by = ?", userID).Find(&helpEvents).WithContext(ctx).Error
	if err != nil {
		return nil, err
	}
	for i := range helpEvents {
		err = h.insertHelpEventMissingData(ctx, &helpEvents[i])
		if err != nil {
			return nil, err
		}
	}

	return helpEvents, nil
}

func (h *HelpEvent) calculatePagination(ctx context.Context, searchValues models.HelpSearchInternal, searchQuery *gorm.DB) (*models.Pagination, error) {
	offset := 0

	if searchValues.Pagination.PageNumber > 0 {
		offset = (searchValues.Pagination.PageNumber - 1) * searchValues.Pagination.PageSize
	}

	events := []models.HelpEvent{}
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

func (h *HelpEvent) UpdateHelpEvent(ctx context.Context, event models.HelpEvent) error {
	if event.File != nil && event.FileType != "" {
		err := h.saveFile(ctx, &event)
		if err != nil {
			return err
		}
	}
	err := h.DB.Model(&event).Updates(event).Where("id = ?", event.ID).WithContext(ctx).Error
	return err
}

func (h *HelpEvent) saveFile(ctx context.Context, event *models.HelpEvent) error {
	if event.File != nil {
		oldEvent := models.HelpEvent{}
		err := h.DB.Where("id = ?", event.ID).First(&oldEvent).WithContext(ctx).Error
		if err != nil {
			return err
		}

		imagePath := strings.Split(oldEvent.ImagePath, "s3.amazonaws.com/")
		imageName := imagePath[len(imagePath)-1]
		err = h.Filer.Delete(ctx, imageName)
		if err != nil {
			return err
		}
		fileName, err := uuid.NewUUID()
		if err != nil {
			return err
		}
		filePath, err := h.Filer.Upload(ctx, fmt.Sprintf("%s.%s", fileName.String(), event.FileType), event.File)
		if err != nil {
			zlog.Log.Error(err, "could not upload file")
			return err
		}
		event.ImagePath = filePath
	}
	return nil
}

func (h *HelpEvent) UpdateNeeds(ctx context.Context, needs ...models.Need) error {
	tx := h.DB.Begin()
	for _, need := range needs {
		err := tx.Model(&need).Updates(need).WithContext(ctx).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (h *HelpEvent) GetHelpEventByTransactionID(ctx context.Context, transactionID models.ID) (models.HelpEvent, error) {
	transaction := models.Transaction{}
	err := h.DB.First(&transaction, "id = ?", transactionID).WithContext(ctx).Error
	if err != nil {
		return models.HelpEvent{}, err
	}

	event := models.HelpEvent{}
	err = h.DB.First(&event, "id = ?", transaction.EventID).WithContext(ctx).Error
	return event, err
}

func (h *HelpEvent) CreateNeed(ctx context.Context, need models.Need) (uint, error) {
	err := h.DB.Create(&need).WithContext(ctx).Error
	return need.ID, err
}

func (h *HelpEvent) GetHelpEventNeeds(ctx context.Context, eventID models.ID) ([]models.Need, error) {
	needs := make([]models.Need, 0)
	err := h.DB.Where("help_event_id = ?", eventID).Where("transaction_id IS NULL").Find(&needs).WithContext(ctx).Error
	return needs, err
}
func (h *HelpEvent) GetEventByID(ctx context.Context, id models.ID) (models.HelpEvent, error) {
	event := models.HelpEvent{
		TransactionNeeds: map[models.ID][]models.Need{},
	}
	err := h.DB.First(&event, id).WithContext(ctx).Error
	if err != nil {
		return models.HelpEvent{}, err
	}

	err = h.insertHelpEventMissingData(ctx, &event)

	return event, err
}

func (h *HelpEvent) insertHelpEventMissingData(ctx context.Context, event *models.HelpEvent) error {
	eventNeeds := make([]models.Need, 0)
	err := h.DB.Where("help_event_id = ?", event.ID).Where("transaction_id IS NULL").Find(&eventNeeds).WithContext(ctx).Error
	if err != nil {
		return err
	}
	event.Needs = eventNeeds
	tags := make([]models.Tag, 0)
	err = h.DB.Where("event_type = ?", models.HelpEventType).Where("event_id = ?", event.ID).Find(&tags).WithContext(ctx).Error
	if err != nil {
		return err
	}
	for i := range tags {
		tagValues := make([]models.TagValue, 0)
		err = h.DB.Where("tag_id = ?", tags[i].ID).Find(&tagValues).WithContext(ctx).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		tags[i].Values = tagValues
	}
	event.Tags = tags
	comments, err := h.getHelpEventComments(ctx, event.ID)
	if err != nil {
		return err
	}
	event.Comments = comments
	if err != nil {
		return err
	}
	user := models.User{}
	err = h.DB.First(&user, "id = ?", event.CreatedBy).WithContext(ctx).Error
	event.User = user
	if err != nil {
		return err
	}

	event.TransactionNeeds = map[models.ID][]models.Need{}
	transactions := make([]models.Transaction, 0)
	err = h.DB.Where("event_type = ?", models.HelpEventType).Where("event_id = ?", event.ID).Find(&transactions).WithContext(ctx).Error
	if err != nil {
		return err
	}
	for i, t := range transactions {
		transactionNeeds := make([]models.Need, 0)
		err = h.DB.Where("transaction_id = ?", t.ID).Find(&transactionNeeds).WithContext(ctx).Error
		if err != nil {
			return err
		}
		event.TransactionNeeds[models.ID(t.ID)] = transactionNeeds
		transactions[i].Needs = transactionNeeds
		transactionCreator := models.User{}
		err = h.DB.Where("id = ?", t.CreatorID).Find(&transactionCreator).WithContext(ctx).Error
		if err != nil {
			return err
		}
		transactions[i].Creator = transactionCreator
	}
	event.Transactions = transactions
	location := models.Address{}
	err = h.DB.
		Where("event_type = ?", models.HelpEventType).
		Where("event_id = ?", event.ID).
		Find(&location).
		WithContext(ctx).
		Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	} else if err == nil {
		event.Location = location
	}

	return nil
}

func (h *HelpEvent) getHelpEventComments(ctx context.Context, eventID uint) ([]models.Comment, error) {
	comments := []models.Comment{}
	err := h.DB.
		Where("event_type = ?", models.HelpEventType).
		Where("event_id = ?", eventID).
		Where("is_deleted = ?", false).
		Find(&comments).
		WithContext(ctx).
		Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return comments, err
	}

	for i, comment := range comments {
		newComment, err := h.updateCommentUsers(ctx, comment)
		if err != nil {
			return nil, err
		}
		comments[i] = newComment
	}

	return comments, nil
}

func (h *HelpEvent) updateCommentUsers(ctx context.Context, comment models.Comment) (models.Comment, error) {
	creatorInfo := models.User{}
	err := h.DB.Where("id = ?", comment.UserID).First(&creatorInfo).WithContext(ctx).Error
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

func (h *HelpEvent) CreateEvent(ctx context.Context, event *models.HelpEvent) (uint, error) {
	tx := h.DB.Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}

	if event.File != nil {
		fileName, err := uuid.NewUUID()
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		filePath, err := h.Filer.Upload(ctx, fmt.Sprintf("%s.%s", fileName.String(), event.FileType), event.File)
		if err != nil {
			zlog.Log.Error(err, "could not upload file")
			return 0, err
		}
		event.ImagePath = filePath
	}

	if err := tx.Create(event).WithContext(ctx).Error; err != nil {
		tx.Rollback()
		return 0, err
	}
	if len(event.Tags) != 0 {
		for i := range event.Tags {
			event.Tags[i].EventID = event.ID
		}
		if err := tx.Create(event.Tags).WithContext(ctx).Error; err != nil {
			tx.Rollback()
			return 0, err
		}
		for i := range event.Tags {
			for j := range event.Tags[i].Values {
				event.Tags[i].Values[j].TagID = event.Tags[i].ID
			}
			if len(event.Tags[i].Values) != 0 {
				if err := tx.Create(event.Tags[i].Values).WithContext(ctx).Error; err != nil {
					tx.Rollback()
					return 0, err
				}
			}
		}
	}

	if !event.Location.IsEmpty() {
		event.Location.EventID = event.ID
		err := tx.
			Create(&event.Location).
			WithContext(ctx).
			Error

		if err != nil {
			tx.Rollback()
			return 0, err
		}
	}

	return event.ID, tx.Commit().Error
}
