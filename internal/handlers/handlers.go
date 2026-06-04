package handlers

import (
	"api/internal/database"
	"api/internal/models"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Handlers struct {
	store *database.SubscriptionStore
}

func NewHandlers(store *database.SubscriptionStore) *Handlers {
	log.Println("[Handlers] Initializing HTTP handlers")
	return &Handlers{store: store}
}

func respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	log.Printf("[Handlers] Error response: status=%d, message=%s", statusCode, message)
	respondWithJSON(w, statusCode, map[string]string{"error": message})
}

// GetAllSubscriptions godoc
// @Summary Get all subscriptions
// @Description Retrieve all subscriptions ordered by start date descending
// @Tags subscriptions
// @Produce json
// @Success 200 {array} models.Subscription
// @Failure 500 {object} map[string]string
// @Router /subscriptions [get]
func (h *Handlers) GetAllSubscriptions(w http.ResponseWriter, r *http.Request) {
	log.Printf("[Handlers.GetAllSubscriptions] Request from %s", r.RemoteAddr)
	start := time.Now()

	subs, err := h.store.GetAll()
	if err != nil {
		log.Printf("[Handlers.GetAllSubscriptions] Error getting subscriptions: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error get subscriptions")
		return
	}

	log.Printf("[Handlers.GetAllSubscriptions] Successfully returned %d subscriptions in %v", len(subs), time.Since(start))
	respondWithJSON(w, http.StatusOK, subs)
}

// GetSubscriptionByID godoc
// @Summary Get subscription by ID
// @Description Retrieve a single subscription by its ID
// @Tags subscriptions
// @Produce json
// @Param id path int true "Subscription ID"
// @Success 200 {object} models.Subscription
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/{id} [get]
func (h *Handlers) GetSubscriptionByID(w http.ResponseWriter, r *http.Request) {
	log.Printf("[Handlers.GetSubscriptionByID] Request from %s", r.RemoteAddr)
	start := time.Now()

	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/subs/"), "/")
	idStr := pathParts[0]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("[Handlers.GetSubscriptionByID] Invalid ID format: %s", idStr)
		respondWithError(w, http.StatusBadRequest, "Incorrect request")
		return
	}

	log.Printf("[Handlers.GetSubscriptionByID] Fetching subscription ID: %d", id)

	sub, err := h.store.GetById(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.Printf("[Handlers.GetSubscriptionByID] Subscription %d not found", id)
			respondWithError(w, http.StatusNotFound, err.Error())
		} else {
			log.Printf("[Handlers.GetSubscriptionByID] Error getting subscription %d: %v", id, err)
			respondWithError(w, http.StatusInternalServerError, "Error get subscription")
		}
		return
	}

	log.Printf("[Handlers.GetSubscriptionByID] Successfully retrieved subscription %d in %v", id, time.Since(start))
	respondWithJSON(w, http.StatusOK, sub)
}

// CreateSubscription godoc
// @Summary Create a new subscription
// @Description Add a new subscription to the database
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body models.CreateSubscriptionInput true "Subscription data"
// @Success 201 {object} models.Subscription
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions [post]
func (h *Handlers) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	log.Printf("[Handlers.CreateSubscription] Request from %s", r.RemoteAddr)
	start := time.Now()

	var input models.CreateSubscriptionInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Printf("[Handlers.CreateSubscription] Failed to decode request body: %v", err)
		respondWithError(w, http.StatusBadRequest, "Incorrect requested data")
		return
	}

	log.Printf("[Handlers.CreateSubscription] Received data: service=%s, price=%d, user=%s, start_date=%s",
		input.ServiceName, input.Price, input.UUID, input.StartDate)

	// Validate required fields
	if strings.TrimSpace(input.UUID) == "" {
		log.Printf("[Handlers.CreateSubscription] Validation failed: UUID is empty")
		respondWithError(w, http.StatusBadRequest, "Incorrect requested data: UUID must be not null")
		return
	}
	if strings.TrimSpace(input.ServiceName) == "" {
		log.Printf("[Handlers.CreateSubscription] Validation failed: service_name is empty")
		respondWithError(w, http.StatusBadRequest, "Incorrect requested data: service name must be not null")
		return
	}
	if strings.TrimSpace(input.StartDate) == "" {
		log.Printf("[Handlers.CreateSubscription] Validation failed: start_date is empty")
		respondWithError(w, http.StatusBadRequest, "Incorrect requested data: start date must be not null")
		return
	}
	if input.Price < 0 {
		log.Printf("[Handlers.CreateSubscription] Validation failed: price %d is negative", input.Price)
		respondWithError(w, http.StatusBadRequest, "Incorrect requested data: price must be positive")
		return
	}

	// Validate date format
	if !isValidMonthYear(input.StartDate) {
		log.Printf("[Handlers.CreateSubscription] Validation failed: invalid start_date format %s", input.StartDate)
		respondWithError(w, http.StatusBadRequest, "start_date must be in MM-YYYY format")
		return
	}
	if input.EndDate != nil && *input.EndDate != "" && !isValidMonthYear(*input.EndDate) {
		log.Printf("[Handlers.CreateSubscription] Validation failed: invalid end_date format %v", input.EndDate)
		respondWithError(w, http.StatusBadRequest, "end_date must be in MM-YYYY format")
		return
	}

	log.Println("[Handlers.CreateSubscription] Validation passed, creating subscription...")

	sub, err := h.store.Create(input)
	if err != nil {
		log.Printf("[Handlers.CreateSubscription] Error creating subscription: %v", err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("[Handlers.CreateSubscription] Successfully created subscription with ID: %v in %v", sub.UUID, time.Since(start))
	respondWithJSON(w, http.StatusCreated, sub)
}

// UpdateSubscription godoc
// @Summary Update a subscription
// @Description Update an existing subscription by ID
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path int true "Subscription ID"
// @Param subscription body models.UpdateSubscriptionInput true "Subscription data to update"
// @Success 200 {object} models.Subscription
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/{id} [put]
func (h *Handlers) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	log.Printf("[Handlers.UpdateSubscription] Request from %s", r.RemoteAddr)
	start := time.Now()

	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/subs/"), "/")
	idStr := pathParts[0]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("[Handlers.UpdateSubscription] Invalid ID format: %s", idStr)
		respondWithError(w, http.StatusBadRequest, "Incorrect request")
		return
	}

	log.Printf("[Handlers.UpdateSubscription] Updating subscription ID: %d", id)

	var input models.UpdateSubscriptionInput
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		log.Printf("[Handlers.UpdateSubscription] Failed to decode request body: %v", err)
		respondWithError(w, http.StatusBadRequest, "Incorrect requested data")
		return
	}

	// Log update fields
	if input.ServiceName != nil {
		log.Printf("[Handlers.UpdateSubscription] Will update service_name to: %s", *input.ServiceName)
	}
	if input.Price != nil {
		log.Printf("[Handlers.UpdateSubscription] Will update price to: %d", *input.Price)
	}
	if input.UUID != nil {
		log.Printf("[Handlers.UpdateSubscription] Will update user_id to: %s", *input.UUID)
	}
	if input.StartDate != nil {
		log.Printf("[Handlers.UpdateSubscription] Will update start_date to: %s", *input.StartDate)
	}
	if input.EndDate != nil {
		if *input.EndDate == "" {
			log.Printf("[Handlers.UpdateSubscription] Will set end_date to NULL")
		} else {
			log.Printf("[Handlers.UpdateSubscription] Will update end_date to: %s", *input.EndDate)
		}
	}

	// Validate fields
	if input.UUID != nil && strings.TrimSpace(*input.UUID) == "" {
		log.Printf("[Handlers.UpdateSubscription] Validation failed: UUID cannot be empty")
		respondWithError(w, http.StatusBadRequest, "Incorrect requested data: UUID must be not null")
		return
	}
	if input.StartDate != nil && strings.TrimSpace(*input.StartDate) == "" {
		log.Printf("[Handlers.UpdateSubscription] Validation failed: start_date cannot be empty")
		respondWithError(w, http.StatusBadRequest, "Incorrect requested data: start date must be not null")
		return
	}
	if input.StartDate != nil && *input.StartDate != "" && !isValidMonthYear(*input.StartDate) {
		log.Printf("[Handlers.UpdateSubscription] Validation failed: invalid start_date format %s", *input.StartDate)
		respondWithError(w, http.StatusBadRequest, "start_date must be in MM-YYYY format")
		return
	}
	if input.EndDate != nil && *input.EndDate != "" && !isValidMonthYear(*input.EndDate) {
		log.Printf("[Handlers.UpdateSubscription] Validation failed: invalid end_date format %s", *input.EndDate)
		respondWithError(w, http.StatusBadRequest, "end_date must be in MM-YYYY format")
		return
	}
	if input.ServiceName != nil && strings.TrimSpace(*input.ServiceName) == "" {
		log.Printf("[Handlers.UpdateSubscription] Validation failed: service_name cannot be empty")
		respondWithError(w, http.StatusBadRequest, "Incorrect requested data: service name must be not null")
		return
	}
	if input.Price != nil && *input.Price < 0 {
		log.Printf("[Handlers.UpdateSubscription] Validation failed: price %d is negative", *input.Price)
		respondWithError(w, http.StatusBadRequest, "Incorrect requested data: price must be positive")
		return
	}

	sub, err := h.store.Update(id, input)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.Printf("[Handlers.UpdateSubscription] Subscription %d not found for update", id)
			respondWithError(w, http.StatusNotFound, err.Error())
		} else {
			log.Printf("[Handlers.UpdateSubscription] Error updating subscription %d: %v", id, err)
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	log.Printf("[Handlers.UpdateSubscription] Successfully updated subscription %d in %v", id, time.Since(start))
	respondWithJSON(w, http.StatusOK, sub)
}

// DeleteSubscription godoc
// @Summary Delete a subscription
// @Description Remove a subscription from the database by ID
// @Tags subscriptions
// @Produce json
// @Param id path int true "Subscription ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/{id} [delete]
func (h *Handlers) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	log.Printf("[Handlers.DeleteSubscription] Request from %s", r.RemoteAddr)
	start := time.Now()

	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/subs/"), "/")
	idStr := pathParts[0]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("[Handlers.DeleteSubscription] Invalid ID format: %s", idStr)
		respondWithError(w, http.StatusBadRequest, "Incorrect request")
		return
	}

	log.Printf("[Handlers.DeleteSubscription] Deleting subscription ID: %d", id)

	err = h.store.Delete(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.Printf("[Handlers.DeleteSubscription] Subscription %d not found for deletion", id)
			respondWithError(w, http.StatusNotFound, err.Error())
		} else {
			log.Printf("[Handlers.DeleteSubscription] Error deleting subscription %d: %v", id, err)
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	log.Printf("[Handlers.DeleteSubscription] Successfully deleted subscription %d in %v", id, time.Since(start))
	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

// GetTotalSpent godoc
// @Summary Calculate total spent on subscriptions
// @Description Get total money spent on subscriptions for a specific period with filtering by user_id and service_name
// @Tags summary
// @Produce json
// @Param user_id query string true "User ID (UUID format)"
// @Param service_name query string true "Service name"
// @Param start_date query string true "Start date in MM-YYYY format"
// @Param end_date query string true "End date in MM-YYYY format"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/summary [get]
func (h *Handlers) GetTotalSpent(w http.ResponseWriter, r *http.Request) {
	log.Printf("[Handlers.GetTotalSpent] Request from %s", r.RemoteAddr)
	start := time.Now()

	userID := r.URL.Query().Get("user_id")
	serviceName := r.URL.Query().Get("service_name")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	log.Printf("[Handlers.GetTotalSpent] Parameters: user=%s, service=%s, start=%s, end=%s",
		userID, serviceName, startDate, endDate)

	// Validate required parameters
	if userID == "" {
		log.Printf("[Handlers.GetTotalSpent] Validation failed: user_id is required")
		respondWithError(w, http.StatusBadRequest, "user_id is required")
		return
	}
	if serviceName == "" {
		log.Printf("[Handlers.GetTotalSpent] Validation failed: service_name is required")
		respondWithError(w, http.StatusBadRequest, "service_name is required")
		return
	}
	if startDate == "" {
		log.Printf("[Handlers.GetTotalSpent] Validation failed: start_date is required")
		respondWithError(w, http.StatusBadRequest, "start_date is required")
		return
	}
	if endDate == "" {
		log.Printf("[Handlers.GetTotalSpent] Validation failed: end_date is required")
		respondWithError(w, http.StatusBadRequest, "end_date is required")
		return
	}

	// Validate date format
	if !isValidMonthYear(startDate) {
		log.Printf("[Handlers.GetTotalSpent] Validation failed: invalid start_date format %s", startDate)
		respondWithError(w, http.StatusBadRequest, "start_date must be in MM-YYYY format")
		return
	}
	if !isValidMonthYear(endDate) {
		log.Printf("[Handlers.GetTotalSpent] Validation failed: invalid end_date format %s", endDate)
		respondWithError(w, http.StatusBadRequest, "end_date must be in MM-YYYY format")
		return
	}

	log.Println("[Handlers.GetTotalSpent] Validation passed, calculating total...")

	total, err := h.store.GetTotalSpent(userID, serviceName, startDate, endDate)
	if err != nil {
		log.Printf("[Handlers.GetTotalSpent] Error calculating total: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error calculating total")
		return
	}

	log.Printf("[Handlers.GetTotalSpent] Successfully calculated total: %d rubles in %v", total, time.Since(start))

	response := map[string]interface{}{
		"user_id":      userID,
		"service_name": serviceName,
		"start_date":   startDate,
		"end_date":     endDate,
		"total_price":  total,
	}
	respondWithJSON(w, http.StatusOK, response)
}

// isValidMonthYear validates date in MM-YYYY format
func isValidMonthYear(date string) bool {
	if len(date) != 7 {
		return false
	}
	if date[2] != '-' {
		return false
	}
	month := date[0:2]
	year := date[3:7]

	if month < "01" || month > "12" {
		return false
	}

	for _, c := range year {
		if c < '0' || c > '9' {
			return false
		}
	}

	return true
}
