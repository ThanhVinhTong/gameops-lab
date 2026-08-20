# AWS adapters

> This package contains AWS SDK for Go v2 client-factory source restricted to LocalStack endpoints. It does not support live AWS endpoints.

## Local-only client policy

- An explicit region and LocalStack endpoint are required.
- Approved endpoints are loopback addresses, `localhost`, `localstack`,
  `host.docker.internal`, and LocalStack's
  `*.localhost.localstack.cloud` names.
- Arbitrary remote hosts, AWS endpoints, URL credentials, paths, queries, and
  fragments are rejected.
- Clients use static dummy credentials, a ten-second HTTP timeout, three
  attempts, no environment proxy, and no redirects.
- Client construction makes no network call, and the factory offers no standard
  AWS fallback.

This fail-closed policy is intentionally unsuitable for live AWS. A separate,
explicit production configuration path would be required if the project later
chooses to use a real AWS account.

This is an application safety guard, not a network-security sandbox. The raw SDK
supports per-operation options, local hostnames still rely on name resolution,
and callers must not override adapter client settings. Context deadlines should
still bound complete operations because retries may make them last longer than
one HTTP timeout.

## EventBridge publisher

`EventBridgePublisher` serializes one `gameops.SessionEnded` event and sends
one `PutEvents` entry to the configured custom bus.

- Source: `gameops.session-api`
- Detail type: `SessionEnded`
- Detail: the domain event's JSON representation

Both operation errors and entry-level EventBridge rejection are returned.

## DynamoDB analytics store

`DynamoDBAnalyticsStore` applies a
`gameops.SessionAnalyticsContribution` using `UpdateItem`.

- Partition key: `USER#<base64url(user ID)>`
- Sort key:
  `GAME#<base64url(game ID)>#PLATFORM#<base64url(platform)>`
- `session_count` and `total_duration_seconds` use atomic numeric additions.
- Descriptive fields and the last processed event fields use last-writer-wins
  assignments.

The encoded key components cannot collide with the `#` delimiters.

## Intentional limitations

- Duplicate delivery increments counters again; idempotency is not implemented.
- An older event processed later can replace the stored last-session timestamp
  and descriptive values.
- No adapter is wired into either executable yet.
- No dependency resolution, compilation, test execution, LocalStack
  connection, or AWS service call has been performed by Codex.
