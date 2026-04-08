#!/bin/bash
set -e

apk add --no-cache curl > /dev/null 2>&1

vault server -dev > /dev/null 2>&1 &
PID=$!

sleep 2

export VAULT_ADDR="http://$VAULT_DEV_LISTEN_ADDRESS"


until curl --silent --fail "$VAULT_ADDR/v1/sys/health" > /dev/null; do
  sleep 1
done

vault auth enable approle


vault policy write "$VAULT_HEALTHCHEK_POLICY" - <<EOF
path "$VAULT_MOUNT_POINT/data/$VAULT_HEALTHCHEK_SECRET_PATH" {
  capabilities = ["read"]
}
EOF

vault policy write "$VAULT_POSTGRES_POLICY" - <<EOF
path "$VAULT_MOUNT_POINT/data/$VAULT_POSTGRES_SECRET_PATH" {
  capabilities = ["read"]
}
EOF

vault write auth/approle/role/$VAULT_HEALTHCHEK_ROLE \
    token_policies="$VAULT_HEALTHCHEK_POLICY" \
    token_ttl="1h" \
    token_max_ttl="4h" \
    secret_id_ttl="60m"

vault write auth/approle/role/$VAULT_POSTGRES_ROLE \
    token_policies="$VAULT_POSTGRES_POLICY" \
    token_ttl="1h" \
    token_max_ttl="4h" \
    secret_id_ttl="60m"


ROLE_ID=$(vault read -field=role_id auth/approle/role/"$VAULT_HEALTHCHEK_ROLE"/role-id)
SECRET_ID=$(vault write -force -field=secret_id auth/approle/role/"$VAULT_HEALTHCHEK_ROLE"/secret-id)

echo "Heathcheck RoleID: $ROLE_ID"
echo "Heathcheck SecretID: $SECRET_ID"

ROLE_ID=$(vault read -field=role_id auth/approle/role/"$VAULT_POSTGRES_ROLE"/role-id)
SECRET_ID=$(vault write -force -field=secret_id auth/approle/role/"$VAULT_POSTGRES_ROLE"/secret-id)

echo "Postgres RoleID: $ROLE_ID"
echo "Postgres SecretID: $SECRET_ID"

vault kv put "$VAULT_MOUNT_POINT/$VAULT_HEALTHCHEK_SECRET_PATH" \
    user="$POSTGRES_USER" \
    password="$POSTGRES_PASSWORD" \
    database="$POSTGRES_DB" \
    host="$POSTGRES_HOST" > /dev/null

vault kv put "$VAULT_MOUNT_POINT/$VAULT_POSTGRES_SECRET_PATH" \
    user="$POSTGRES_USER" \
    password="$POSTGRES_PASSWORD" \
    database="$POSTGRES_DB" \
    host="$POSTGRES_HOST" > /dev/null

echo "Vault успешно инициализирован!"

wait $PID