package models

type Subscription struct {
	ServiceName string `json:"service_name" db:"service_name"`
	Price       int64  `json:"price" db:"price"`
	UUID        string `json:"user_id" db:"user_id"`
	StartDate   string `json:"start_date" db:"start_date"`
}

type CreateSubscriptionInput struct {
	ServiceName string `json:"service_name"`
	Price       int64  `json:"price"`
	UUID        string `json:"user_id"`
	StartDate   string `json:"start_date"`
}

type UpdateSubscriptionInput struct {
	ServiceName *string `json:"service_name"`
	Price       *int64  `json:"price"`
	UUID        *string `json:"user_id"`
	StartDate   *string `json:"start_date"`
}
