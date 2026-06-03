package handlers

import (
	"api/internal/database"
	"api/internal/models"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type Handlers struct {
	store *database.SubscriptionStore
}

func NewHandlers(store *database.SubscriptionStore) *Handlers {
	return &Handlers{store: store}
}

func respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	respondWithJSON(w, statusCode, map[string]string{"error": message})
}

func (h *Handlers) GetAllSubscriptions(w http.ResponseWriter, r *http.Request) {
	subs, err := h.store.GetAll()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error get subscriptions")
		return
	}
	respondWithJSON(w, http.StatusOK, subs)
}
func (h *Handlers) GetSubscriptionByID(w http.ResponseWriter, r *http.Request) {
	pathPatrs := strings.Split(strings.TrimPrefix(r.URL.Path, "/subs/"), "/")
	idStr := pathPatrs[0]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Incorrect request")
		return
	}
	subs, err := h.store.GetById(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error get subscription")
		return
	}
	respondWithJSON(w, http.StatusOK, subs)
}
func (h *Handlers) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var input models.CreateSubscriptionInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Incorrect requested data")
		return
	}
	if strings.TrimSpace(input.UUID) == "" {
		respondWithError(w, http.StatusBadRequest, "Incorrect requested data: UUID must be not null")
		return
	}
	if strings.TrimSpace(input.ServiceName) == "" {
		respondWithError(w, http.StatusBadRequest, "Incorrect requested data: service name must be not null")
		return
	}
	if strings.TrimSpace(input.StartDate) == "" {
		respondWithError(w, http.StatusBadRequest, "Incorrect requested data: start date must be not null")
		return
	}
	if input.Price < 0 {
		respondWithError(w, http.StatusBadRequest, "Incorrect requested data: price must be positive")
		return
	}
	sub, err := h.store.Create(input)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusCreated, sub)
}
func (h *Handlers) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	pathPatrs := strings.Split(strings.TrimPrefix(r.URL.Path, "/subs/"), "/")
	idStr := pathPatrs[0]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Incorrect request")
		return
	}
	var input models.UpdateSubscriptionInput
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Incorrect requested data")
		return
	}
	if input.UUID != nil && strings.TrimSpace(*input.UUID) == "" {
		respondWithError(w, http.StatusBadRequest, "Incorrect requested data: UUID must be not null")
		return
	}
	if input.StartDate != nil && strings.TrimSpace(*input.StartDate) == "" {
		respondWithError(w, http.StatusBadRequest, "Incorrect requested data: start date must be not null")
		return
	}
	if input.ServiceName != nil && strings.TrimSpace(*input.ServiceName) == "" {
		respondWithError(w, http.StatusBadRequest, "Incorrect requested data: service name must be not null")
		return
	}
	if input.Price != nil && *input.Price < 0 {
		respondWithError(w, http.StatusBadRequest, "Incorrect requested data: price must be positive")
		return
	}
	sub, err := h.store.Update(id, input)
	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			respondWithError(w, http.StatusNotFound, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())

		}
	}
	respondWithJSON(w, http.StatusOK, sub)
}

func (h *Handlers) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	pathPatrs := strings.Split(strings.TrimPrefix(r.URL.Path, "/subs/"), "/")
	idStr := pathPatrs[0]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Incorrect request")
		return
	}
	err = h.store.Delete(id)
	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			respondWithError(w, http.StatusNotFound, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())

		}
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})

}
func (h *Handlers) GetTotalSpent(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	serviceName := r.URL.Query().Get("service_name")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	if userID == "" {
		respondWithError(w, http.StatusBadRequest, "user_id is required")
		return
	}
	if serviceName == "" {
		respondWithError(w, http.StatusBadRequest, "service_name is required")
		return
	}
	if startDate == "" {
		respondWithError(w, http.StatusBadRequest, "start_date is required")
		return
	}
	if endDate == "" {
		respondWithError(w, http.StatusBadRequest, "end_date is required")
		return
	}

	// Validate date format
	if !isValidMonthYear(startDate) {
		respondWithError(w, http.StatusBadRequest, "start_date must be in MM-YYYY format")
		return
	}
	if !isValidMonthYear(endDate) {
		respondWithError(w, http.StatusBadRequest, "end_date must be in MM-YYYY format")
		return
	}

	total, err := h.store.GetTotalSpent(userID, serviceName, startDate, endDate)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error calculating total")
		return
	}

	response := map[string]interface{}{
		"user_id":      userID,
		"service_name": serviceName,
		"start_date":   startDate,
		"end_date":     endDate,
		"total_price":  total,
	}
	respondWithJSON(w, http.StatusOK, response)
}

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
