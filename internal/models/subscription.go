package models

// Subscription represents a user's subscription to a service
type Subscription struct {
	ID          int     `json:"id" db:"id" example:"1"`
	ServiceName string  `json:"service_name" db:"service_name" example:"Netflix"`
	Price       int64   `json:"price" db:"price" example:"599"`
	UUID        string  `json:"user_id" db:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string  `json:"start_date" db:"start_date" example:"01-2024"`
	EndDate     *string `json:"end_date,omitempty" db:"end_date" example:"12-2024"`
}

// CreateSubscriptionInput represents the data needed to create a new subscription
type CreateSubscriptionInput struct {
	ServiceName string  `json:"service_name" binding:"required" example:"Yandex Plus"`
	Price       int64   `json:"price" binding:"required" example:"400"`
	UUID        string  `json:"user_id" binding:"required" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string  `json:"start_date" binding:"required" example:"07-2025"`
	EndDate     *string `json:"end_date,omitempty" example:"12-2025"`
}

// UpdateSubscriptionInput represents the data that can be updated for a subscription
type UpdateSubscriptionInput struct {
	ServiceName *string `json:"service_name,omitempty" example:"Netflix Premium"`
	Price       *int64  `json:"price,omitempty" example:"649"`
	UUID        *string `json:"user_id,omitempty" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   *string `json:"start_date,omitempty" example:"01-2024"`
	EndDate     *string `json:"end_date,omitempty" example:"12-2024"`
}
