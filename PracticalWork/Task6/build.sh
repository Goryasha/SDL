#!/bin/bash
set -e

cd "$(dirname "$0")"

source build.env

ROLE_ID=$VAULT_POSTGRES_ROLE_ID
SECRET_ID=$VAULT_POSTGRES_SECRET_ID

TOKEN=$(curl -sS --request POST --data "{\"role_id\":\"$ROLE_ID\",\"secret_id\":\"$SECRET_ID\"}" "$VAULT_ADDR/v1/auth/approle/login" | jq -r .auth.client_token)
POSTGRES_USER=$(curl -sS -H "X-Vault-Token: $TOKEN" $VAULT_ADDR/v1/$VAULT_MOUNT_POINT/data/$VAULT_POSTGRES_SECRET_PATH | jq -r .data.data.user)
POSTGRES_PASSWORD=$(curl -sS -H "X-Vault-Token: $TOKEN" $VAULT_ADDR/v1/$VAULT_MOUNT_POINT/data/$VAULT_POSTGRES_SECRET_PATH | jq -r .data.data.password)
POSTGRES_DB=$(curl -sS -H "X-Vault-Token: $TOKEN" $VAULT_ADDR/v1/$VAULT_MOUNT_POINT/data/$VAULT_POSTGRES_SECRET_PATH | jq -r .data.data.database)
HOST=$(curl -sS -H "X-Vault-Token: $TOKEN" $VAULT_ADDR/v1/$VAULT_MOUNT_POINT/data/$VAULT_POSTGRES_SECRET_PATH | jq -r .data.data.host)

sed -i -E "s/^(POSTGRES_USER)=.*/\1=$POSTGRES_USER/" "DB_scripts/.env"
sed -i -E "s/^(POSTGRES_PASSWORD)=.*/\1=$POSTGRES_PASSWORD/" "DB_scripts/.env"
sed -i -E "s/^(POSTGRES_DB)=.*/\1=$POSTGRES_DB/" "DB_scripts/.env"
sed -i -E "s/^(HOST)=.*/\1=$HOST/" "DB_scripts/.env"

sed "s/POSTGRES_USER/$POSTGRES_USER/g" ./DB_scripts/init.sql > ./DB_scripts/init_filled.sql

docker compose build --no-cache
docker compose up --renew-anon-volumes --remove-orphans &
PID=$!

sleep 10
rm -r ./DB_scripts/init_filled.sql

sed -i -E "s/^(POSTGRES_USER)=.*/\1=POSTGRES_USER/" "DB_scripts/.env"
sed -i -E "s/^(POSTGRES_PASSWORD)=.*/\1=POSTGRES_PASSWORD/" "DB_scripts/.env"
sed -i -E "s/^(POSTGRES_DB)=.*/\1=POSTGRES_DB/" "DB_scripts/.env"
sed -i -E "s/^(HOST)=.*/\1=HOST/" "DB_scripts/.env"

wait $PID
