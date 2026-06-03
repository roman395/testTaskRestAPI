package database

import (
	"api/internal/models"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type SubscriptionStore struct {
	db *sqlx.DB
}

func NewSubscriptionStore(db *sqlx.DB) *SubscriptionStore {
	return &SubscriptionStore{db: db}
}

func (s *SubscriptionStore) GetAll() ([]models.Subscription, error) {
	var subs []models.Subscription

	query :=
		`Select id, service_name, price, user_id, start_date, end_date 
	FROM subscriptions
	order by start_date desc`
	err := s.db.Select(&subs, query)
	if err != nil {
		return nil, err
	}
	return subs, nil
}

func (s *SubscriptionStore) GetById(id int) (*models.Subscription, error) {
	var sub models.Subscription

	query :=
		`Select id, service_name, price, user_id, start_date, end_date 
	FROM subscriptions
	where id = $1`

	err := s.db.Get(&sub, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf(`subscription with id %d not found`, id)
	}
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (s *SubscriptionStore) Create(input models.CreateSubscriptionInput) (*models.Subscription, error) {
	var sub models.Subscription

	query :=
		`INSERT INTO subscriptions(service_name, price, user_id, start_date, end_date) 
		VALUES ($1, $2, $3, $4, $5)
		returning id, service_name, price, user_id, start_date, end_date;
		`
	err := s.db.QueryRowx(query, input.ServiceName, input.Price, input.UUID, input.StartDate, input.EndDate).StructScan(&sub)
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (s *SubscriptionStore) Update(id int, input models.UpdateSubscriptionInput) (*models.Subscription, error) {
	sub, err := s.GetById(id)
	if err != nil {
		return nil, err
	}
	if input.ServiceName != nil {
		sub.ServiceName = *input.ServiceName
	}
	if input.Price != nil {
		sub.Price = *input.Price
	}
	if input.UUID != nil {
		sub.UUID = *input.UUID
	}
	if input.StartDate != nil {
		sub.StartDate = *input.StartDate
	}
	if input.EndDate != nil {
		sub.EndDate = *input.EndDate
	}

	query := `
	UPDATE subscriptions
	SET service_name = $1, price = $2, user_id = $3, start_date = $4, end_date = $5
	where id = $6
	returning id, service_name, price, user_id, start_date, end_date`
	var updatedSub models.Subscription

	err = s.db.QueryRowx(query, sub.ServiceName, sub.Price, sub.UUID, sub.StartDate, sub.EndDate, id).StructScan(&updatedSub)
	if err != nil {
		return nil, err
	}
	return &updatedSub, err
}

func (s *SubscriptionStore) Delete(id int) error {
	query := `DELETE FROM subscriptions WHERE id = $1`
	result, err := s.db.Exec(query, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf(`subscription with id %d not found`, id)
	}
	return nil
}

func (s *SubscriptionStore) GetTotalSpent(userID, serviceName, startDate, endDate string) (int, error) {
	var total int

	query := `
        SELECT COALESCE(SUM(price), 0)
        FROM subscriptions
        WHERE user_id = $1 
        AND service_name = $2
        AND TO_DATE(start_date, 'MM-YYYY') >= TO_DATE($3, 'MM-YYYY')
        AND (end_date IS NULL OR TO_DATE(end_date, 'MM-YYYY') <= TO_DATE($4, 'MM-YYYY'))
    `

	err := s.db.Get(&total, query, userID, serviceName, startDate, endDate)
	if err != nil {
		return 0, err
	}
	return total, nil
}
