#!/bin/bash

curl -X PUT http://localhost:8091/api/orders \
  -H "Content-Type: application/json" \
  -d '{
    "id": 13,
    "status": "completed"  }'

echo