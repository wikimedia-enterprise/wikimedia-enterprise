
From this directory, run `docker-compose up --build` to start Auth plus dependencies, including a Cognito emulator. Users user1, user2 and user3 should be available, belonging to groups group_1, group_2 and group_3, with passwords HardPass1!, HardPass2! and HardPass3!.

Place an .env file in the root folder of the repository, as usual:

```
AWS_REGION=us-east-1
AWS_ID=some_id
AWS_KEY=some_secret
COGNITO_SECRET=your_secret
SERVER_MODE=release
SERVER_PORT=4050
COGNITO_USER_GROUP=group_3 # for create-user
REDIS_ADDR=redis:6379
REDIS_PASSWORD=password
ACCESS_POLICY=policy-csv
```

Requests to `/v1/login` should work. The successful responses will include `"challenge_name": "PASSWORD_VERIFIER"`, but they will also have the necessary tokens and no further action is required.
