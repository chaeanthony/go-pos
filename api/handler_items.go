package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/chaeanthony/go-pos/internal/database"
	"github.com/chaeanthony/go-pos/utils"
	"github.com/shopspring/decimal"
)

func (cfg *APIConfig) HandlerItemsGet(w http.ResponseWriter, r *http.Request) {
	type ItemResponse struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Cost        string `json:"cost"`
		CreatedAt   string `json:"created_at"`
		UpdatedAt   string `json:"updated_at"`
	}

	items, err := cfg.DB.GetItems()
	if err != nil {
		utils.RespondError(w, cfg.Logger, http.StatusInternalServerError, "Couldn't get items", err)
		return
	}

	// Map items to response items with cost as decimal (float64)
	respItems := make([]ItemResponse, len(items))
	for i, item := range items {
		costDecimal := float64(item.Cost) / 100.0
		respItems[i] = ItemResponse{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			Cost:        fmt.Sprintf("%.2f", costDecimal), // format with 2 decimals as string
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		}
	}

	utils.RespondJSON(w, cfg.Logger, http.StatusOK, respItems)
}

func (cfg *APIConfig) HandlerItemGetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("itemID")
	if id == "" {
		utils.RespondError(w, cfg.Logger, http.StatusBadRequest, "Missing itemID", fmt.Errorf("missing itemID"))
		return
	}
	item, err := cfg.DB.GetItemByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.RespondError(w, cfg.Logger, http.StatusNotFound, "Couldn't find item", err)
		} else {
			utils.RespondError(w, cfg.Logger, http.StatusInternalServerError, "Couldn't get item", err)
		}
		return
	}

	utils.RespondJSON(w, cfg.Logger, http.StatusOK, item)
}

func (cfg *APIConfig) HandlerItemsCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		Cost        decimal.Decimal `json:"cost"`
	}

	params := parameters{}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		utils.RespondError(w, cfg.Logger, http.StatusBadRequest, "Couldn't decode parameters", err)
		return
	}

	itemID, err := cfg.DB.CreateItem(database.CreateItemParams{
		Name:        params.Name,
		Description: params.Description,
		Cost:        int(params.Cost.Mul(decimal.NewFromInt(100)).IntPart()),
	})
	if err != nil {
		utils.RespondError(w, cfg.Logger, http.StatusInternalServerError, "Couldn't create item", err)
		return
	}

	utils.RespondJSON(w, cfg.Logger, http.StatusCreated, map[string]string{"id": strconv.Itoa(int(itemID))})
}

func (cfg *APIConfig) HandlerItemsUpdate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		ID          string          `json:"id"`
		Name        string          `json:"name"`
		Description string          `json:"description"`
		Cost        decimal.Decimal `json:"cost"`
	}

	params := parameters{}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		utils.RespondError(w, cfg.Logger, http.StatusBadRequest, "Couldn't decode parameters", err)
		return
	}

	err = cfg.DB.UpdateItem(database.UpdateItemParams{
		ID: params.ID, Name: params.Name, Description: params.Description, Cost: int(params.Cost.Mul(decimal.NewFromInt(100)).IntPart()),
	})
	if err != nil {
		utils.RespondError(w, cfg.Logger, http.StatusInternalServerError, "Couldn't update item", err)
		return
	}

	utils.RespondJSON(w, cfg.Logger, http.StatusOK, map[string]string{"status": "updated", "id": params.ID})
}

func (cfg *APIConfig) HandlerItemsDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("itemID")

	if id == "" {
		utils.RespondError(w, cfg.Logger, http.StatusBadRequest, "Missing itemID", nil)
		return
	}

	_, err := cfg.DB.GetItemByID(id)
	if err != nil {
		utils.RespondError(w, cfg.Logger, http.StatusNotFound, "Couldn't find item", err)
		return
	}

	err = cfg.DB.DeleteItem(id)
	if err != nil {
		utils.RespondError(w, cfg.Logger, http.StatusInternalServerError, "Couldn't delete item", err)
		return
	}

	utils.RespondJSON(w, cfg.Logger, http.StatusOK, map[string]string{"status": "deleted", "id": id})
}
