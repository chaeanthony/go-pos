-- +goose Up
CREATE TABLE IF NOT EXISTS orders(
  id INTEGER PRIMARY KEY,
  created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  for_name TEXT NOT NULL,
  for_email TEXT NOT NULL,
  order_date TEXT,
  status TEXT NOT NULL,
  total TEXT NOT NULL DEFAULT "0.00",
  notes TEXT
);

-- +goose Down
DROP TABLE orders;