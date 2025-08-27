#!/bin/bash

# Reset env vars, to prevent SSO failures
export AWS_ACCESS_KEY_ID=local
export AWS_SECRET_ACCESS_KEY=local
export AWS_REGION=us-nowhere-1

CMD="aws cognito-idp  --endpoint-url ${COGNITO_ENDPOINT:?}"
until $CMD list-user-pools --max-results 1; do
    echo "Waiting for cognito-local to be ready at ${COGNITO_ENDPOINT}..."
    sleep 2
done
echo "cognito-local is ready!"

USER_POOL_ID=$($CMD create-user-pool --pool-name all_users --query 'UserPool.Id' --output text)
$CMD update-user-pool --user-pool-id $USER_POOL_ID --mfa-configuration OFF
# If needed:
# $CMD describe-user-pool --user-pool-id $USER_POOL_ID
echo "Created user pool with ID: $USER_POOL_ID"

# Create App Client ID
TMPF=$(mktemp)
$CMD create-user-pool-client --user-pool-id $USER_POOL_ID --client-name LocalCognitoAppClient --generate-secret > $TMPF
COGNITO_CLIENT_ID=$(jq -r .UserPoolClient.ClientId < $TMPF)
COGNITO_CLIENT_SECRET=$(jq -r .UserPoolClient.ClientSecret < $TMPF)
rm $TMPF

# Create groups
groups=(group_1 group_2 group_3)
for group in "${groups[@]}"; do
    $CMD create-group --user-pool-id $USER_POOL_ID --group-name $group
    echo "Created group: $group"
done

# Create users and add to groups
$CMD admin-create-user --user-pool-id $USER_POOL_ID --username user1 --user-attributes Name=email,Value=user1@example.com --temporary-password TempPass1!  --message-action SUPPRESS
$CMD admin-set-user-password --user-pool-id $USER_POOL_ID --username user1 --password 'HardPass1!' --permanent
$CMD admin-add-user-to-group --user-pool-id $USER_POOL_ID --username user1 --group-name group_1

$CMD admin-create-user --user-pool-id $USER_POOL_ID --username user2 --user-attributes Name=email,Value=user2@example.com --temporary-password TempPass2! --message-action SUPPRESS
$CMD admin-set-user-password --user-pool-id $USER_POOL_ID --username user2 --password 'HardPass2!' --permanent
$CMD admin-add-user-to-group --user-pool-id $USER_POOL_ID --username user2 --group-name group_2

$CMD admin-create-user --user-pool-id $USER_POOL_ID --username user3 --user-attributes Name=email,Value=user3@example.com --temporary-password TempPass3! --message-action SUPPRESS
$CMD admin-set-user-password --user-pool-id $USER_POOL_ID --username user3 --password 'HardPass3!' --permanent
$CMD admin-add-user-to-group --user-pool-id $USER_POOL_ID --username user3 --group-name group_3
echo $COGNITO_CLIENT_ID > "${CLIENT_ID_FILE:?}"
echo $COGNITO_CLIENT_SECRET > "${CLIENT_SECRET_FILE:?}"
echo $USER_POOL_ID > "${POOL_ID_FILE:?}"

echo "Setup complete!"
