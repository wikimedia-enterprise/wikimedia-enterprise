#!/bin/bash

# Wait until the files are created
while [ ! -f "${POOL_ID_FILE:?}" ]; do
  echo "Waiting for Cognito setup to complete..."
  sleep 3
done

export COGNITO_CLIENT_ID=$(cat "$CLIENT_ID_FILE")
export COGNITO_CLIENT_SECRET=$(cat "$CLIENT_SECRET_FILE")
export COGNITO_USER_POOL_ID=$(cat "$POOL_ID_FILE")
echo "Found client ID:" $COGNITO_CLIENT_ID
echo "Found client secret:" $COGNITO_CLIENT_SECRET
echo "Found user pool ID:" $COGNITO_USER_POOL_ID

# Start the service
./main
