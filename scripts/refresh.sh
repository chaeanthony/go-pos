#!/bin/bash

curl -X POST http://localhost:8091/api/refresh \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoic3RvcmUiLCJpc3MiOiJnby1ob21lLXBvcyIsInN1YiI6IjQxMDVjNzBiLTRkNGUtNGFiYy04MTI5LWYxMDkzZmFiMGRmOSIsImV4cCI6MTc0OTIzMDUxNSwiaWF0IjoxNzQ5MjI5NjE1fQ.H8MAFPxcelvz4BjvpiVTm4pLobBOWqeX3OfiMNrT69A" 

echo