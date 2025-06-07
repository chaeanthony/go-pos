#!/bin/bash

curl -X DELETE http://localhost:8091/api/items/0196de5f-3733-4be1-93e3-f1cc5e722101 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiYWRtaW4iLCJpc3MiOiJnby1ob21lLXBvcyIsInN1YiI6IjI5ZWMwMjA4LThjNTAtNDcwNC1iYTlmLTI2YWM4MjlhOGY3NSIsImV4cCI6MTc0NDgzODYzMCwiaWF0IjoxNzQ0ODM1MDMwfQ.f_uwK8UNbY8NwaJO37sQXejKKRikbajiLgdQmlC3bak" 

echo