package database

import (
	"github.com/google/uuid"
)

type Item struct {
	ID string `json:"id"`
	CreateItemParams
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type CreateItemParams struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Cost        int    `json:"cost"`
}

type UpdateItemParams struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Cost        int    `json:"cost"`
}

func (c *Client) GetItems() ([]Item, error) {
	query := `
	SELECT id, name, description, cost, created_at, updated_at FROM items
	`

	rows, err := c.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []Item{}
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Cost, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (c *Client) GetItemByID(id string) (*Item, error) {
	query := `
	SELECT id, name, description, cost, created_at, updated_at FROM items WHERE id = ?
	`

	var item Item
	err := c.db.QueryRow(query, id).Scan(&item.ID, &item.Name, &item.Description, &item.Cost, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, err // also err if error is sql.ErrNoRows
	}

	return &item, nil
}

func (c *Client) CreateItem(params CreateItemParams) (int64, error) {
	query := `
	INSERT INTO items (id, name, description, cost, created_at, updated_at)
	VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	id := uuid.NewString()
	res, err := c.db.Exec(query, id, params.Name, params.Description, params.Cost)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (c *Client) UpdateItem(params UpdateItemParams) error {
	query := `
	UPDATE items SET name = ?, description = ?, cost = ?, updated_at = CURRENT_TIMESTAMP
	WHERE id = ?
	`

	_, err := c.db.Exec(query, params.Name, params.Description, params.Cost, params.ID)

	return err
}

func (c *Client) DeleteItem(id string) error {
	query := `
	DELETE FROM items WHERE id = ?
	`

	_, err := c.db.Exec(query, id)
	if err != nil {
		return err
	}

	return nil
}
