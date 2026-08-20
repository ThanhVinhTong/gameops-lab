# GAMEOPS LAB — CODEX IMPLEMENTATION PROMPTS

Use all prompts in the same Codex conversation and in the listed order.

---

# PROMPT 0 — PERMANENT SESSION RULES

You are working on the `gameops-lab` repository.

## Strict execution policy

You are allowed to:

* Read existing repository files.
* Create, edit, move, and delete project files when necessary.
* Explain code and architectural decisions.
* Provide commands for me to run manually.

You are not allowed to:

* Execute shell, PowerShell, Bash, Git, Go, Docker, AWS CLI, SAM CLI, Python, Node.js, package-manager, build, test, formatting, linting, deployment, or network commands.
* Install packages or tools.
* Start containers or processes.
* Access AWS or deploy any resource to real AWS.
* Perform Git operations, create commits, change branches, or push code.
* Claim that code builds, tests pass, containers start, or the application works unless I later provide the execution output.
* Rewrite unrelated existing code.
* Add a frontend, authentication system, Steam integration, Kubernetes, Terraform, or other functionality outside the requested scope.
* Ask for confirmation before making ordinary implementation decisions.

Do not run commands even when doing so appears useful. Make code changes only.

## Working behaviour

Before editing:

1. Read the relevant existing files.
2. Preserve useful existing code and repository conventions.
3. Identify the smallest coherent set of changes required by the current prompt.

While editing:

1. Use idiomatic Go.
2. Use the existing Go module and Go version where available.
3. Use AWS SDK for Go v2.
4. Use standard-library functionality unless an external dependency provides clear value.
5. Keep AWS endpoints configurable.
6. Never hardcode real credentials, account IDs, secrets, tokens, queue URLs, table names, event-bus names, or local endpoints into domain logic.
7. Keep handlers thin and business logic testable through interfaces.
8. Use structured JSON logging with `log/slog`.
9. Wrap errors with useful context.
10. Do not leave TODO placeholders for core functionality requested in the current prompt.
11. Do not silently replace existing files with incomplete stubs.
12. Do not introduce abstractions that have only one trivial use unless they improve testing or local/cloud portability.

## Local-only safety requirement

This project must run locally without creating billable AWS resources.

Local development will use:

* Docker Compose.
* LocalStack Hobby unified image.
* AWS SAM.
* Dummy local AWS credentials.
* A configurable local AWS endpoint.

Do not describe LocalStack as an actively maintained open-source Community edition.

Do not add any automatic path that deploys to real AWS.

Every AWS CLI command placed in documentation or scripts for local resources must include:

```text
--endpoint-url http://localhost:4566
```

Every local SAM deployment command must use `samlocal`, not `sam deploy`.

The application must still be portable to production AWS by leaving the endpoint configuration empty.

## Required response after each prompt

After editing, return only:

1. `Summary`
2. `Files changed`
3. `Important implementation decisions`
4. `Commands for me to run manually`
5. `Unverified assumptions or risks`

Explicitly state:

> No commands were executed, so the build and tests have not been verified.

Acknowledge these rules without editing code yet.

---

# PROMPT 1 — FOUNDATION, LOCAL ENVIRONMENT, AND AWS SAM INFRASTRUCTURE

Implement the project foundation and local infrastructure for GameOps Lab.

Do not execute any commands.

## Product goal

GameOps Lab is a local-first, AWS-compatible event-driven platform written in Go. It receives completed gaming sessions and asynchronously calculates weekly gaming analytics.

The target event flow is:

```text
Client
  |
  v
API Gateway HTTP API
  |
  v
Session API Lambda
  |
  v
Custom EventBridge Bus
  |
  v
EventBridge Rule
  |
  v
SQS Analytics Queue ------> SQS Dead-Letter Queue
  |
  v
Analytics Consumer Lambda
  |
  v
DynamoDB
```

The entire system must be runnable locally without provisioning real AWS resources.

## Scope of this prompt

Implement only:

* Repository foundation.
* Configuration package.
* AWS client construction.
* Docker Compose.
* AWS SAM infrastructure.
* Empty Lambda entry points that compile structurally.
* Local PowerShell helper scripts.
* Initial README instructions.

Do not implement the complete session API or analytics processing yet.

## Expected repository structure

Adapt this structure to existing files rather than deleting useful work:

```text
gameops-lab/
├── cmd/
│   ├── session-api/
│   │   └── main.go
│   └── analytics-consumer/
│       └── main.go
├── internal/
│   ├── config/
│   ├── awsclient/
│   ├── events/
│   ├── sessions/
│   ├── analytics/
│   ├── storage/
│   └── idempotency/
├── scripts/
│   ├── local-up.ps1
│   ├── local-deploy.ps1
│   ├── local-status.ps1
│   └── local-down.ps1
├── docs/
│   └── adr/
├── template.yaml
├── docker-compose.yml
├── env.local.json.example
├── .env.example
├── .gitignore
└── README.md
```

Keep empty directories only when they will be used in later prompts. Prefer small package documentation files over meaningless `.gitkeep` files.

## Configuration requirements

Create a configuration package that reads, validates, and exposes at least:

```text
AWS_REGION
AWS_ENDPOINT_URL
EVENT_BUS_NAME
DYNAMODB_TABLE_NAME
LOG_LEVEL
IDEMPOTENCY_TTL_DAYS
```

Requirements:

* Default region: `ap-southeast-1`.
* Local endpoint may be empty.
* Empty endpoint means production AWS SDK behaviour.
* Do not hardcode `http://localstack:4566` in domain or repository code.
* Configuration errors must explain which value is invalid.
* Table name and event-bus name must be supplied through Lambda environment variables.
* Parse log level into a `slog.Level`.

## AWS clients

Create AWS SDK for Go v2 constructors for:

* EventBridge.
* DynamoDB.

Requirements:

* Load the default AWS configuration.
* Honour `AWS_REGION`.
* If `AWS_ENDPOINT_URL` is non-empty, configure the clients to use it.
* If it is empty, use normal AWS endpoints.
* Do not embed credentials.
* Return contextual errors.
* Keep the client constructors reusable and testable.

## Docker Compose

Create or update `docker-compose.yml`.

Use:

```yaml
image: ${LOCALSTACK_IMAGE:-localstack/localstack-pro:latest}
```

Requirements:

* Service name and container hostname: `localstack`.
* Container name: `gameops-localstack`.
* Port `4566:4566`.
* Named Docker network: `gameops-local`.
* Mount `/var/run/docker.sock`.
* Read `LOCALSTACK_AUTH_TOKEN` from `.env`.
* Do not commit an actual token.
* Configure region `ap-southeast-1`.
* Enable only the services required by this project:

```text
lambda
apigateway
events
sqs
dynamodb
cloudformation
iam
logs
sts
```

* Configure Lambda containers to use the `gameops-local` network.
* Do not enable persistence unless the existing repository already has a known working configuration.
* Do not add a real AWS account dependency.

Create `.env.example` with safe placeholders:

```text
LOCALSTACK_AUTH_TOKEN=
LOCALSTACK_IMAGE=localstack/localstack-pro:latest
LOCALSTACK_DEBUG=0
AWS_DEFAULT_REGION=ap-southeast-1
```

Add `.env`, LocalStack runtime data, `.aws-sam`, generated binaries, coverage files, IDE files, and secret files to `.gitignore`.

## AWS SAM template

Create or update `template.yaml` using AWS SAM and CloudFormation.

Define:

### HTTP API

Use `AWS::Serverless::HttpApi`.

Initially expose:

```text
POST /sessions/end
GET /users/{userId}/weekly/{week}
GET /health
```

The GET weekly route may remain functionally unimplemented until a later prompt, but the route should be connected to the session API Lambda.

### Session API Lambda

* Go custom runtime using `provided.al2023`.
* Build method compatible with Go SAM builds.
* Handler/bootstrap structure suitable for Go.
* Timeout around 15 seconds.
* Memory around 256 MB.
* Environment variables for region, endpoint, event bus, table, log level, and idempotency TTL.
* Least-privilege permission to call `events:PutEvents` on the custom event bus.
* Read-only DynamoDB permission for the weekly summary endpoint.

### Custom EventBridge bus

Create a named custom event bus.

### SQS analytics queue

Create:

* Analytics queue.
* Analytics DLQ.
* Redrive policy with `maxReceiveCount: 3`.
* Visibility timeout safely greater than the analytics Lambda timeout.
* Long polling where appropriate.

### EventBridge rule

Match:

```json
{
  "source": ["gameops.session"],
  "detail-type": ["SessionEnded"]
}
```

Target the analytics SQS queue.

Add the queue policy required for EventBridge to send messages to the queue. Restrict the policy to the rule ARN rather than granting unrestricted SQS access.

### Analytics consumer Lambda

* Go custom runtime using `provided.al2023`.
* Triggered by the analytics SQS queue.
* Batch size no greater than 10.
* Enable partial batch failure reporting with `ReportBatchItemFailures`.
* Timeout around 30 seconds.
* Memory around 256 MB.
* Least-privilege DynamoDB write permissions.
* Environment variables from the shared configuration.
* Do not give EventBridge permissions to this consumer.

### DynamoDB table

Use one table with:

```text
PK: string
SK: string
```

Use:

* On-demand billing.
* TTL attribute named `expires_at`.
* Server-side encryption.
* Point-in-time recovery enabled for the production target template when supported.
* Retain or deletion behaviour should be explicitly documented.

Do not add unnecessary GSIs for the MVP.

### Outputs

Output at least:

* HTTP API URL.
* Event bus name.
* Analytics queue URL.
* Analytics DLQ URL.
* DynamoDB table name.

## Lambda entry points

Create minimal entry points for both Lambdas.

For now they may return clearly marked “not implemented” behaviour, but they must:

* Load configuration.
* Set up structured logging.
* Construct the required AWS clients.
* Avoid package-global mutable clients where practical.
* Avoid claiming the actual handlers are finished.

## Local environment example

Create `env.local.json.example` containing dummy credentials and:

```text
AWS_ENDPOINT_URL=http://localstack:4566
AWS_REGION=ap-southeast-1
```

Do not put a real LocalStack token in this file.

## PowerShell scripts

Create safe PowerShell scripts intended for Windows.

### `local-up.ps1`

It should contain commands for the user to:

* Check `.env` exists.
* Start Docker Compose.
* Show container status.
* Check LocalStack health.

### `local-deploy.ps1`

It should contain commands for the user to:

* Build using `sam build`.
* Deploy only to LocalStack using `samlocal deploy`.
* Never call `sam deploy`.
* Use stack name `gameops-lab-local`.
* Use region `ap-southeast-1`.
* Print CloudFormation outputs through AWS CLI with the LocalStack endpoint.

### `local-status.ps1`

It should contain endpoint-safe commands to show:

* CloudFormation stack.
* SQS queues.
* Event buses.
* DynamoDB tables.
* LocalStack container status.

### `local-down.ps1`

It should stop Docker Compose without deleting source files.

Scripts must:

* Use `$ErrorActionPreference = "Stop"`.
* Avoid modifying global AWS profiles.
* Avoid storing credentials.
* Clearly display what they are doing.
* Never contain a command targeting real AWS.

## README foundation

Document:

* Project goal.
* Architecture.
* Why the project is local-first.
* Prerequisites.
* LocalStack Hobby authentication requirement.
* Local setup steps.
* Explicit commands that the user must run.
* Safety warning not to use `sam deploy`.
* Local-to-production service mapping.
* Current implementation status.
* A statement that local emulators do not guarantee complete AWS behavioural parity.

Use Mermaid for the initial architecture diagram.

Do not claim the application is complete or tested.

## Acceptance criteria for this prompt

The repository should contain a coherent foundation for later implementation, but you must not run or verify it.

At the end, follow the required response format from Prompt 0.

---

# PROMPT 2 — SESSION API, EVENT CONTRACT, AND WEEKLY READ ENDPOINT

Implement the GameOps Lab session API and event publishing flow.

Do not execute any commands.

Read and preserve the infrastructure and conventions created previously.

## Scope

Implement:

```text
POST /sessions/end
GET /users/{userId}/weekly/{week}
GET /health
```

The POST endpoint publishes a `SessionEnded` event to the custom EventBridge bus.

The GET weekly endpoint reads already-generated analytics from DynamoDB.

Do not implement the SQS analytics consumer in this prompt.

## Domain model

Create a versioned event envelope.

Suggested logical structure:

```json
{
  "version": 1,
  "event_id": "UUID",
  "event_type": "SessionEnded",
  "occurred_at": "RFC3339 timestamp",
  "data": {
    "user_id": "user-123",
    "game_id": "elden-ring",
    "game_name": "Elden Ring",
    "platform": "pc",
    "started_at": "RFC3339 timestamp",
    "ended_at": "RFC3339 timestamp",
    "duration_seconds": 7200
  }
}
```

The HTTP request may omit:

```text
event_id
duration_seconds
```

The API must:

* Preserve a valid client-provided event ID so duplicate-event scenarios can be demonstrated.
* Generate a cryptographically random UUID-compatible ID when absent.
* Calculate duration from `started_at` and `ended_at`.
* Reject a provided duration if it conflicts materially with the calculated duration.
* Set `occurred_at` on the server.
* Use UTC internally.

Avoid a UUID dependency if a correct implementation can be produced using `crypto/rand`.

## Validation

Validate at least:

* JSON body is present and valid.
* Unknown JSON fields are rejected.
* Required string fields are not blank.
* IDs have safe length limits.
* Platform is one of a small documented set such as:

```text
pc
playstation
xbox
switch
mobile
other
```

* Timestamps are valid RFC3339 values.
* `ended_at` is later than `started_at`.
* Session duration is positive.
* Session duration is no longer than 24 hours for the MVP.
* Future timestamps beyond a small clock-skew tolerance are rejected.
* Event ID has a safe format and length.

Return structured error responses:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "ended_at must be later than started_at",
    "request_id": "..."
  }
}
```

Do not expose stack traces or internal AWS errors to clients.

## POST response

On successful EventBridge publication, return HTTP `202 Accepted`:

```json
{
  "event_id": "...",
  "status": "accepted",
  "request_id": "..."
}
```

The API must not write the session directly to DynamoDB.

Reason:

* Avoid a dual-write between DynamoDB and EventBridge.
* Event delivery and persistence belong to the asynchronous consumer.
* Document this decision in an ADR or README architecture section.

## EventBridge publishing

Use AWS SDK for Go v2.

Publish:

```text
Source: gameops.session
DetailType: SessionEnded
EventBusName: configured custom bus
Detail: event envelope JSON
```

Requirements:

* Check every `PutEvents` result entry.
* Treat `FailedEntryCount > 0` as an error.
* Log AWS error code and message safely.
* Include request ID and event ID in logs.
* Return `503 Service Unavailable` when EventBridge publication fails.
* Do not log the entire request payload by default.
* Do not retry indefinitely inside the API Lambda.

Create a publisher interface so the handler can be unit tested without AWS.

## API Gateway adapter

Implement a thin API Gateway HTTP API Lambda adapter.

Requirements:

* Route based on HTTP method and raw path or route key.
* Propagate or create a request/correlation ID.
* Set JSON content type.
* Handle CORS only if already needed; do not add permissive wildcard CORS unnecessarily.
* Keep HTTP concerns separate from domain validation and EventBridge publishing.

## Health endpoint

`GET /health` should return:

```json
{
  "status": "ok",
  "service": "session-api"
}
```

It should not make an AWS call.

## Weekly analytics read endpoint

Implement:

```text
GET /users/{userId}/weekly/{week}
```

Expected week format:

```text
YYYY-Www
```

Example:

```text
2026-W32
```

Validate both user ID and week.

Use a DynamoDB repository interface.

Expected response:

```json
{
  "user_id": "user-123",
  "week": "2026-W32",
  "total_seconds": 10800,
  "session_count": 2,
  "games": [
    {
      "game_id": "elden-ring",
      "total_seconds": 7200,
      "session_count": 1
    }
  ]
}
```

Read items using:

```text
PK = USER#<userId>
SK begins_with SUMMARY#<week>
```

Expected item forms:

```text
PK = USER#<userId>
SK = SUMMARY#<week>

PK = USER#<userId>
SK = SUMMARY#<week>#GAME#<gameId>
```

Requirements:

* Use consistent reads where practical for the local demo.
* Return `404` when no weekly summary exists.
* Sort games deterministically by total time descending, then game ID ascending.
* Do not scan the table.
* Do not expose DynamoDB attribute structures to the HTTP layer.

## Logging

Use `log/slog` JSON logs containing useful fields such as:

```text
service
request_id
event_id
user_id
operation
status
duration_ms
```

Do not log credentials, LocalStack token, raw authorization headers, or the complete event payload.

## Tests

Add focused unit tests without introducing a heavy mock framework.

Test at least:

* Valid session request.
* Invalid JSON.
* Unknown JSON field.
* End before start.
* Session longer than 24 hours.
* Invalid platform.
* Client-provided event ID is preserved.
* Missing event ID is generated.
* EventBridge publisher failure maps to 503.
* EventBridge partial failure maps to 503.
* Successful request maps to 202.
* Weekly key validation.
* Weekly response aggregation and deterministic game ordering.
* Health response.

Do not run the tests.

## Documentation

Update the README with:

* API contract.
* Example valid request.
* Example accepted response.
* Validation rules.
* Explanation of why the API publishes an event rather than writing DynamoDB directly.
* Manual local invocation commands that target local endpoints only.

Update the project status accurately.

At the end, follow the required response format from Prompt 0.

---

# PROMPT 3 — SQS CONSUMER, ATOMIC IDEMPOTENCY, RETRIES, AND DLQ

Implement the analytics SQS consumer and DynamoDB transaction.

Do not execute any commands.

Read the existing event contract rather than creating a conflicting second definition.

## Consumer input

The analytics Lambda receives SQS messages created by an EventBridge rule.

The SQS body contains an EventBridge event wrapper whose `detail` contains the versioned GameOps event envelope.

The consumer must:

1. Parse each SQS record independently.
2. Parse the EventBridge wrapper.
3. Parse and validate the `SessionEnded` detail.
4. Reject unsupported event versions.
5. Reject unexpected event type, source, or detail type.
6. Derive the ISO week from the session end timestamp in UTC.
7. Persist the event through one atomic DynamoDB transaction.
8. Return partial batch failures for records that should be retried.

## DynamoDB data model

Use these items.

### Idempotency item

```text
PK = EVENT#<eventId>
SK = IDEMPOTENCY
entity_type = IDEMPOTENCY
event_id
processed_at
expires_at
```

`expires_at` is a Unix timestamp derived from configured `IDEMPOTENCY_TTL_DAYS`.

### Session item

```text
PK = USER#<userId>
SK = SESSION#<endedAtUTC>#<eventId>
entity_type = SESSION
event_id
user_id
game_id
game_name
platform
started_at
ended_at
duration_seconds
created_at
```

Use a stable sortable RFC3339 UTC representation in the sort key.

### Weekly summary item

```text
PK = USER#<userId>
SK = SUMMARY#<YYYY-Www>
entity_type = WEEKLY_SUMMARY
week
total_seconds
session_count
updated_at
```

### Weekly game summary item

```text
PK = USER#<userId>
SK = SUMMARY#<YYYY-Www>#GAME#<gameId>
entity_type = WEEKLY_GAME_SUMMARY
week
game_id
game_name
total_seconds
session_count
updated_at
```

## Atomic transaction

Use one DynamoDB `TransactWriteItems` operation containing:

1. Conditional put for the idempotency item:

```text
attribute_not_exists(PK)
```

2. Put for the session item with an appropriate non-overwrite condition.

3. Update weekly summary:

```text
ADD total_seconds :duration, session_count :one
SET entity_type = if_not_exists(...),
    week = if_not_exists(...),
    updated_at = :now
```

4. Update weekly game summary using the same pattern.

The idempotency check and aggregate updates must be in the same transaction.

Do not implement this sequence as:

```text
check
write
update
```

with separate calls, because that permits races and partial state.

## Duplicate handling

When the transaction is cancelled because the idempotency conditional check failed:

* Treat the message as successfully processed.
* Do not increment analytics again.
* Log it as a duplicate.
* Do not return it in `BatchItemFailures`.

Do not classify every transaction cancellation as a duplicate.

Inspect cancellation reasons or otherwise implement a reliable classification so capacity errors, validation errors, and service failures are retried.

Create a small domain-level result such as:

```text
Processed
Duplicate
RetryableFailure
PermanentFailure
```

where helpful.

## Retry and DLQ behaviour

Use Lambda SQS partial batch response:

```json
{
  "batchItemFailures": [
    {
      "itemIdentifier": "sqs-message-id"
    }
  ]
}
```

Classify failures:

### Retryable

Examples:

* DynamoDB service error.
* Throttling.
* Temporary AWS SDK error.
* Transaction conflict.
* Internal unexpected error.

Return these record IDs as batch failures.

### Permanent malformed event

Malformed or semantically invalid messages should also be returned as failed so SQS redrive moves them to the DLQ after `maxReceiveCount`.

Log them with a clear reason, but do not panic.

Do not silently acknowledge malformed records.

### Duplicate

A known duplicate is successful and must not be retried.

## Lambda handler quality

Requirements:

* Process records independently.
* Do not fail the complete batch when only one record fails.
* Recover safely from a panic at the record-processing boundary and mark that record failed.
* Respect context cancellation.
* Keep AWS-specific handler code separate from analytics processing.
* Use structured logs with message ID, event ID, user ID, attempt metadata where available, and result.
* Do not log the full message body for malformed messages; log a bounded diagnostic.
* Do not use goroutines for the first implementation.
* Do not manually delete SQS messages.
* Do not implement custom sleep or retry loops inside the Lambda.

## Tests

Add focused unit tests for:

* Valid EventBridge-wrapped SQS message.
* Invalid SQS body.
* Invalid EventBridge wrapper.
* Invalid event detail.
* Unsupported event version.
* Unexpected event source.
* Duplicate result is acknowledged.
* Retryable DynamoDB error returns the message ID in batch failures.
* Malformed message returns the message ID in batch failures.
* Mixed batch returns only failed record IDs.
* Panic in one record does not prevent processing the other records.
* ISO week calculation across year boundaries.
* DynamoDB transaction contains all four expected operations.
* Idempotency conditional expression exists.
* Summary updates use atomic `ADD`.
* Duplicate transaction cancellation is classified correctly.
* Non-duplicate cancellation remains retryable.

Avoid brittle tests tied to irrelevant implementation details.

Do not run tests.

## SAM verification by inspection

Inspect and adjust `template.yaml` if necessary so that:

* SQS partial batch response is enabled.
* DLQ redrive uses `maxReceiveCount: 3`.
* Queue visibility timeout is safely above Lambda timeout.
* Consumer has only required DynamoDB permissions.
* The DynamoDB table has TTL configured on `expires_at`.
* EventBridge has only the permission needed to send to the target queue.

Do not deploy or run validation commands.

## Documentation

Update README with a failure-handling section explaining:

* At-least-once delivery.
* Why duplicate delivery is expected.
* Atomic idempotency.
* Partial batch failures.
* Retry path.
* DLQ path.
* Why malformed messages are not acknowledged.
* What LocalStack can and cannot prove about production AWS behaviour.

Add a Mermaid sequence diagram for:

```text
Client → API → EventBridge → SQS → Consumer → DynamoDB
```

Add a second diagram showing retry and DLQ behaviour.

At the end, follow the required response format from Prompt 0.

---

# PROMPT 4 — LOCAL DEMO, DOCUMENTATION, ADRS, AND FINAL CODE REVIEW

Complete the portfolio-ready local demo and review the repository.

Do not execute any commands.

Do not add a frontend.

## Objective

Make the repository understandable and manually runnable by a recruiter or engineer on Windows with:

* Docker Desktop.
* Go.
* AWS CLI.
* AWS SAM CLI.
* `samlocal`.
* LocalStack Hobby credentials.

All execution remains local.

## PowerShell smoke test

Create or improve:

```text
scripts/smoke-test.ps1
```

The script should be readable, safe, and fail clearly.

It must only contain commands for the user to run manually.

It should:

1. Confirm LocalStack is healthy.
2. Read CloudFormation outputs from:

```text
gameops-lab-local
```

3. Obtain the local API endpoint, queue URL, DLQ URL, and table name without hardcoding generated values.
4. Submit a valid gaming session with a fixed event ID.
5. Submit the same event ID again.
6. Poll the weekly analytics endpoint until the summary appears or a bounded timeout is reached.
7. Verify that `session_count` increased only once.
8. Display the resulting weekly summary.
9. Send a deliberately malformed message directly to the analytics queue.
10. Explain that SQS must receive it multiple times before it appears in the DLQ.
11. Optionally poll the DLQ with a bounded timeout.
12. Never call AWS without:

```text
--endpoint-url http://localhost:4566
```

13. Never contain real credentials.
14. Never perform a real AWS deployment.

Use dummy local environment variables only inside script process scope.

The script should not depend on `jq`; use PowerShell JSON support.

## Manual failure demo

Create:

```text
scripts/demo-failure-recovery.ps1
```

Document or implement commands for the user to:

1. Identify the analytics Lambda event source mapping.
2. Disable processing locally or otherwise pause the consumer using the safest supported local method.
3. Submit several valid sessions.
4. Show messages accumulating in SQS.
5. Restore processing.
6. Show messages being consumed.
7. Re-submit an existing event.
8. Show that analytics are not double-counted.
9. Submit malformed data.
10. Show eventual DLQ routing.

If reliably pausing the consumer is not possible with the chosen emulator, do not invent a command. Instead document a safe alternative demo, such as temporarily disabling the event source mapping through LocalStack’s AWS endpoint.

Clearly mark emulator-specific limitations.

## Documentation structure

Create or update:

```text
docs/
├── architecture.md
├── production-target.md
├── threat-model.md
├── demo-guide.md
└── adr/
    ├── 0001-local-first-development.md
    ├── 0002-eventbridge-and-sqs.md
    ├── 0003-dynamodb-over-postgresql.md
    └── 0004-lambda-over-containers.md
```

## Architecture documentation

`architecture.md` must explain:

* Components and responsibilities.
* Event contract.
* DynamoDB access patterns.
* At-least-once delivery.
* Idempotency boundary.
* Retry and DLQ behaviour.
* Why the API does not directly write session state.
* Local-to-AWS mapping.
* LocalStack limitations.
* Which guarantees are verified by code and which require validation on AWS.

Include:

* Local architecture Mermaid diagram.
* Production target Mermaid diagram.
* Request sequence.
* Failure sequence.

## Production target

`production-target.md` should describe, without deploying:

* API Gateway HTTP API.
* Lambda.
* EventBridge.
* SQS and DLQ.
* DynamoDB.
* CloudWatch logs and alarms.
* IAM least privilege.
* Encryption in transit and at rest.
* DynamoDB backups and point-in-time recovery.
* SQS DLQ alarms.
* Lambda error/throttle alarms.
* API throttling.
* Structured observability.
* Correlation IDs.
* Deployment environments.
* How configuration changes when `AWS_ENDPOINT_URL` becomes empty.
* Likely scaling characteristics.
* Cost drivers rather than fake exact prices.
* Migration path from one user to a larger workload.
* Differences between local emulation and AWS-managed behaviour.

Do not invent benchmark numbers or claim production validation.

## Threat model

Keep it practical and concise.

Cover:

* Malicious or malformed payloads.
* Oversized payloads.
* Replay and duplicate events.
* Unauthorized event publication.
* Overly broad IAM permissions.
* Secret leakage.
* Log leakage.
* Queue poisoning.
* Unbounded retry.
* Denial of service and cost amplification.
* Local Docker socket risk.
* LocalStack token handling.

For each risk include:

```text
risk
impact
current mitigation
production recommendation
```

## ADR requirements

Each ADR should contain:

```text
Status
Context
Decision
Alternatives considered
Consequences
When to revisit
```

### ADR 0001

Explain local-first development as a zero-cost customer constraint.

### ADR 0002

Explain EventBridge plus SQS rather than direct synchronous analytics processing.

Mention:

* Loose coupling.
* Buffering.
* Independent retries.
* DLQ.
* Added operational complexity.

### ADR 0003

Explain DynamoDB versus PostgreSQL.

Mention:

* Known access patterns.
* Conditional writes.
* Atomic transactions.
* Serverless operational model.
* Harder ad hoc querying.
* Data duplication.
* Potential hot-key concerns.

Do not claim DynamoDB is universally better.

### ADR 0004

Explain Lambda versus ECS/container services.

Mention:

* Bursty workload.
* Event integration.
* Reduced operations.
* Execution limits.
* Cold starts.
* When sustained workloads would justify containers.

## README final form

The root README should be recruiter-friendly and concise.

Include:

1. Project statement.
2. Why it exists.
3. Architecture diagram.
4. Engineering highlights.
5. Technology stack.
6. Quick start.
7. API examples.
8. Demo scenarios.
9. Data model.
10. Reliability design.
11. Local versus production mapping.
12. Current limitations.
13. Repository structure.
14. Documentation links.
15. Honest project status.

Use this framing:

> Built and tested an AWS-compatible event-driven architecture locally using AWS SAM and LocalStack Hobby without provisioning billable AWS resources.

Do not say:

> Deployed to AWS

unless the repository actually contains evidence of an AWS deployment, which is outside this task.

## Final code review

Review all project files by inspection.

Fix issues involving:

* Conflicting event definitions.
* Hardcoded endpoints.
* Hardcoded generated resource names.
* Missing context propagation.
* Unwrapped errors.
* Incorrect HTTP status mapping.
* DynamoDB key inconsistencies.
* Non-atomic idempotency.
* Incomplete partial-batch failure handling.
* Permissive IAM.
* Missing EventBridge-to-SQS queue policy.
* Unsafe logs.
* Secret files accidentally included.
* Documentation commands that could target real AWS.
* Unnecessary dependencies.
* Dead or duplicate code.
* Misleading claims.
* Placeholder TODOs in core functionality.

Do not perform unrelated refactoring.

## Final test inventory

List the exact commands I should run manually, in order, but do not execute them:

```text
go mod tidy
gofmt
go vet
go test
sam validate
sam build
docker compose config
docker compose up
samlocal deploy
smoke-test.ps1
```

Use correct concrete syntax in the final response and README.

Separate:

* Commands that only inspect or test code.
* Commands that start the local environment.
* Commands that deploy to LocalStack.
* Commands that perform the demo.
* Commands that stop the environment.

Every AWS CLI command for the local environment must contain the LocalStack endpoint.

At the end, follow the required response format from Prompt 0.

---

# OPTIONAL PROMPT 5 — FIX ISSUES FROM MY MANUAL EXECUTION

I ran the commands manually and will paste the complete output below.

Follow the permanent rules from Prompt 0.

Do not run any command yourself.

Tasks:

1. Diagnose errors only from the repository and the output I provide.
2. Distinguish root causes from secondary errors.
3. Make the smallest coherent code or configuration changes needed.
4. Do not bypass failures by deleting tests, weakening validation, disabling security controls, or removing the event-driven flow.
5. Do not claim the issue is fixed until I run the commands again.
6. Tell me exactly which commands I should rerun manually.

Here is the complete output:

```text
PASTE OUTPUT HERE
```
