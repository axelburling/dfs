#!/bin/sh

set -a
[ -f .env ] && . .env

echo "Using database source: $DATABASE_URL"

if [[ $1 == 'up' ]]
then
    migrate -database "$DATABASE_URL" -path ./pkg/db/migrations/ up
elif [[ $1 == 'down' ]]
then
    migrate -database "$DATABASE_URL" -path ./pkg/db/migrations/ down
elif [[ $1 == 'drop' ]]
then
    migrate -database "$DATABASE_URL" -path ./pkg/db/migrations/ drop
else
    read -p "Run your own command: " custom_command
    migrate -database "$DATABASE_URL" -path ./pkg/db/migrations/ $custom_command 
fi

