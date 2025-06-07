package database

import (
	"errors"
	"fmt"
)

type CreateOrderParams struct {
	ForName   string                  `json:"for_name"`
	ForEmail  string                  `json:"for_email"`
	OrderDate string                  `json:"order_date"`
	Status    string                  `json:"status"`
	Total     string                  `json:"total"` // Stored as a string to handle decimal formatting
	Notes     string                  `json:"notes"` // Additional notes for the order
	Items     []CreateOrderItemParams `json:"items"` // Associated order items
}

type CreateOrderItemParams struct {
	ItemID   string `json:"item_id"`
	Quantity int    `json:"quantity"`
	Price    string `json:"price"`
	Notes    string `json:"notes"` // Additional notes for the order item
}

type UpdateOrderParams struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
}

var ErrOrderNotFound = errors.New("order not found")

func (c *Client) GetOrdersJSON() (string, error) {
	// query formatted like so to return JSON ordered by order_date
	query := `SELECT json_group_array(json(order_json)) AS orders_json
		FROM (
			SELECT
				'{
					"id": ' || o.id || ',
					"for_name": ' || json_quote(o.for_name) || ',
					"email": ' || json_quote(o.for_email) || ',
					"order_date": ' || json_quote(o.order_date) || ',
					"status": ' || json_quote(o.status) || ',
					"total": ' || o.total || ',
					"created_at": ' || json_quote(o.created_at) || ',
					"updated_at": ' || json_quote(o.updated_at) || ',
					"items": ' || (
						SELECT COALESCE(json_group_array(
							json_object(
								'id', oi.id,
								'order_id', oi.order_id,
								'item_name', i.name,
								'item_description', i.description,
								'quantity', oi.quantity,
								'price', oi.price,
								'notes', oi.notes
							)
						), '[]')
						FROM order_items oi
						JOIN items i ON oi.item_id = i.id
						WHERE oi.order_id = o.id
						ORDER BY oi.id
					) || '
				}' AS order_json
			FROM orders o
			WHERE o.status != "completed"
			ORDER BY o.order_date ASC
		); `

	var ordersJSON string
	err := c.db.QueryRow(query).Scan(&ordersJSON)

	return ordersJSON, err
}

func (c *Client) CreateOrder(order CreateOrderParams) (int, error) {
	// Begin a transaction
	tx, err := c.db.Begin()
	if err != nil {
		return 0, err
	}

	// Insert the order and get its ID
	orderQuery := `
		INSERT INTO orders (for_name, for_email, order_date, status, total, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id
	`

	var orderID int
	err = tx.QueryRow(orderQuery,
		order.ForName,
		order.ForEmail,
		order.OrderDate,
		order.Status,
		order.Total,
	).Scan(&orderID)
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("failed to create order: %v", err)
	}

	// Insert order items
	itemQuery := `
		INSERT INTO order_items (order_id, item_id, quantity, price, notes)
		VALUES (?, ?, ?, ?, ?)
	`

	stmt, err := tx.Prepare(itemQuery)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	defer stmt.Close()

	for _, item := range order.Items {
		_, err := stmt.Exec(orderID, item.ItemID, item.Quantity, item.Price, item.Notes)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("failed to create order items: %v", err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return orderID, nil
}

func (c *Client) UpdateOrder(order UpdateOrderParams) error {
	query := `
		UPDATE orders
		SET status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := c.db.Exec(query, order.Status, order.ID)
	if err != nil {
		return fmt.Errorf("failed to update order: %v", err)
	}

	return nil
}

// DeleteOrder removes an order by ID
func (c *Client) DeleteOrder(id int) error {
	query := `DELETE FROM orders WHERE id = ?`
	_, err := c.db.Exec(query, id)
	return err
}

// OrderExists checks if an order exists with orderID
func (c *Client) OrderExists(orderID int) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM orders WHERE id = ?)"
	err := c.db.QueryRow(query, orderID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
