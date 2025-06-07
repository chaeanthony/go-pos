#!/bin/bash

curl -X POST http://localhost:8091/api/orders \
  -H "Content-Type: application/json" \
  --data @order.json

echo
