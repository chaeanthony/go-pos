#!/bin/bash

curl -b cookies.txt -X POST http://localhost:8091/api/items \
  -H "Content-Type: application/json" \
  -d '{"name":"Espresso 3","description":"Black espresso. Chocolate with hint of plum","cost":6.00}'

echo