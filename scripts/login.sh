#!/bin/bash

curl -c cookies.txt -X POST http://localhost:8091/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"chaeanthony21@gmail.com", "password":"yourepretty"}'

echo