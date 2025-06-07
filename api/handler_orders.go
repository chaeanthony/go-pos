package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/chaeanthony/go-pos/internal/database"
	"github.com/chaeanthony/go-pos/utils"
)

func (cfg *APIConfig) HandlerOrdersGet(w http.ResponseWriter, r *http.Request) {
	ordersJSON, err := cfg.DB.GetOrdersJSON()
	if err != nil {
		utils.RespondError(w, cfg.Logger, http.StatusInternalServerError, "Couldn't get orders", err)
		return
	}

	utils.Respond(w, cfg.Logger, http.StatusOK, ordersJSON)
}

func (cfg *APIConfig) HandlerOrdersCreate(w http.ResponseWriter, r *http.Request) {
	params := database.CreateOrderParams{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		utils.RespondError(w, cfg.Logger, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	_, err := time.Parse(database.TIME_LAYOUT, params.OrderDate)
	if err != nil {
		utils.RespondError(w, cfg.Logger, http.StatusBadRequest, "Invalid order date format", err)
		return
	}

	id, err := cfg.DB.CreateOrder(params)
	if err != nil {
		utils.RespondError(w, cfg.Logger, http.StatusInternalServerError, "Couldn't create order", err)
		return
	}

	utils.RespondJSON(w, cfg.Logger, http.StatusCreated, map[string]string{"id": strconv.Itoa(id)})
	cfg.broadcastRefreshOrders()
}

func (cfg *APIConfig) HandlerOrdersUpdate(w http.ResponseWriter, r *http.Request) {
	params := database.UpdateOrderParams{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		utils.RespondError(w, cfg.Logger, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	err := cfg.DB.UpdateOrder(params)
	if err != nil {
		utils.RespondError(w, cfg.Logger, http.StatusInternalServerError, "Couldn't update order", err)
		return
	}

	utils.RespondJSON(w, cfg.Logger, http.StatusOK, map[string]string{"message": fmt.Sprintf("Order %d updated successfully", params.ID)})
	cfg.broadcastRefreshOrders()
}

func (cfg *APIConfig) broadcastRefreshOrders() {
	// Notify all clients to refresh their orders
	msg, err := json.Marshal(struct {
		Type string `json:"type"`
	}{
		Type: "refresh_orders",
	})
	if err != nil {
		cfg.Logger.Errorf("couldn't marshal refresh message: %v", err)
		return
	}
	cfg.Hub.Broadcast(msg)
}
