package database

import (
	"api/internal/models"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

type SubscriptionStore struct {
	db *sqlx.DB
}

func NewSubscriptionStore(db *sqlx.DB) *SubscriptionStore {
	log.Println("[SubscriptionStore] Initializing new store")
	return &SubscriptionStore{db: db}
}

func (s *SubscriptionStore) GetAll() ([]models.Subscription, error) {
	log.Println("[SubscriptionStore.GetAll] Fetching all subscriptions")
	start := time.Now()

	var subs []models.Subscription
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date 
		FROM subscriptions
		ORDER BY start_date DESC
	`

	err := s.db.Select(&subs, query)
	if err != nil {
		log.Printf("[SubscriptionStore.GetAll] Error fetching subscriptions: %v", err)
		return nil, err
	}

	log.Printf("[SubscriptionStore.GetAll] Successfully retrieved %d subscriptions in %v", len(subs), time.Since(start))
	return subs, nil
}

func (s *SubscriptionStore) GetById(id int) (*models.Subscription, error) {
	log.Printf("[SubscriptionStore.GetById] Fetching subscription with ID: %d", id)
	start := time.Now()

	var sub models.Subscription
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date 
		FROM subscriptions
		WHERE id = $1
	`

	err := s.db.Get(&sub, query, id)
	if err == sql.ErrNoRows {
		log.Printf("[SubscriptionStore.GetById] Subscription with ID %d not found", id)
		return nil, fmt.Errorf("subscription with id %d not found", id)
	}
	if err != nil {
		log.Printf("[SubscriptionStore.GetById] Error fetching subscription %d: %v", id, err)
		return nil, err
	}

	log.Printf("[SubscriptionStore.GetById] Successfully retrieved subscription %d in %v", id, time.Since(start))
	return &sub, nil
}

func (s *SubscriptionStore) Create(input models.CreateSubscriptionInput) (*models.Subscription, error) {
	log.Printf("[SubscriptionStore.Create] Creating new subscription: service=%s, user=%s, price=%d",
		input.ServiceName, input.UUID, input.Price)
	start := time.Now()

	var sub models.Subscription
	query := `
		INSERT INTO subscriptions(service_name, price, user_id, start_date, end_date) 
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, service_name, price, user_id, start_date, end_date
	`

	err := s.db.QueryRowx(query, input.ServiceName, input.Price, input.UUID, input.StartDate, input.EndDate).StructScan(&sub)
	if err != nil {
		log.Printf("[SubscriptionStore.Create] Error creating subscription: %v", err)
		return nil, err
	}

	log.Printf("[SubscriptionStore.Create] Successfully created subscription with ID: %v in %v", sub.UUID, time.Since(start))
	return &sub, nil
}

func (s *SubscriptionStore) Update(id int, input models.UpdateSubscriptionInput) (*models.Subscription, error) {
	log.Printf("[SubscriptionStore.Update] Updating subscription ID: %d", id)
	start := time.Now()

	sub, err := s.GetById(id)
	if err != nil {
		log.Printf("[SubscriptionStore.Update] Cannot update subscription %d: not found", id)
		return nil, err
	}

	// Log changes being made
	if input.ServiceName != nil {
		log.Printf("[SubscriptionStore.Update] Changing service_name from '%s' to '%s'", sub.ServiceName, *input.ServiceName)
		sub.ServiceName = *input.ServiceName
	}
	if input.Price != nil {
		log.Printf("[SubscriptionStore.Update] Changing price from %d to %d", sub.Price, *input.Price)
		sub.Price = *input.Price
	}
	if input.UUID != nil {
		log.Printf("[SubscriptionStore.Update] Changing user_id from '%s' to '%s'", sub.UUID, *input.UUID)
		sub.UUID = *input.UUID
	}
	if input.StartDate != nil {
		log.Printf("[SubscriptionStore.Update] Changing start_date from '%s' to '%s'", sub.StartDate, *input.StartDate)
		sub.StartDate = *input.StartDate
	}
	if input.EndDate != nil {
		if sub.EndDate == "" {
			log.Printf("[SubscriptionStore.Update] Setting end_date from NULL to '%s'", *input.EndDate)
		} else {
			log.Printf("[SubscriptionStore.Update] Changing end_date from '%s' to '%s'", sub.EndDate, *input.EndDate)
		}
		sub.EndDate = *input.EndDate
	}

	query := `
		UPDATE subscriptions
		SET service_name = $1, price = $2, user_id = $3, start_date = $4, end_date = $5
		WHERE id = $6
		RETURNING id, service_name, price, user_id, start_date, end_date
	`

	var updatedSub models.Subscription
	err = s.db.QueryRowx(query, sub.ServiceName, sub.Price, sub.UUID, sub.StartDate, sub.EndDate, id).StructScan(&updatedSub)
	if err != nil {
		log.Printf("[SubscriptionStore.Update] Error updating subscription %d: %v", id, err)
		return nil, err
	}

	log.Printf("[SubscriptionStore.Update] Successfully updated subscription %d in %v", id, time.Since(start))
	return &updatedSub, nil
}

func (s *SubscriptionStore) Delete(id int) error {
	log.Printf("[SubscriptionStore.Delete] Deleting subscription ID: %d", id)
	start := time.Now()

	query := `DELETE FROM subscriptions WHERE id = $1`
	result, err := s.db.Exec(query, id)
	if err != nil {
		log.Printf("[SubscriptionStore.Delete] Error deleting subscription %d: %v", id, err)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		log.Printf("[SubscriptionStore.Delete] Error getting affected rows for subscription %d: %v", id, err)
		return err
	}

	if rows == 0 {
		log.Printf("[SubscriptionStore.Delete] Subscription with ID %d not found", id)
		return fmt.Errorf("subscription with id %d not found", id)
	}

	log.Printf("[SubscriptionStore.Delete] Successfully deleted subscription %d in %v", id, time.Since(start))
	return nil
}

func (s *SubscriptionStore) GetTotalSpent(userID, serviceName, startDate, endDate string) (int, error) {
	log.Printf("[SubscriptionStore.GetTotalSpent] Calculating total for user=%s, service=%s, period=%s to %s",
		userID, serviceName, startDate, endDate)
	start := time.Now()

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
		log.Printf("[SubscriptionStore.GetTotalSpent] Error calculating total: %v", err)
		return 0, err
	}

	log.Printf("[SubscriptionStore.GetTotalSpent] Total calculated: %d rubles in %v", total, time.Since(start))
	return total, nil
}
