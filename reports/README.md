# Reports and runtime evidence

This directory will preserve reviewable evidence that the local-only project
behaves as described. Reports should record the command or action, expected
behavior, observed behavior, limitations, and the real runtime evidence that
supports the result.

- `templates/` contains reusable Markdown templates.
- `assets/` contains real runtime screenshots.
- Future completed Markdown reports live directly in `reports/`.

To create a report, copy a template into this directory, give it a descriptive
name, replace every `{{PLACEHOLDER}}`, run the relevant verification only at
the appropriate development stage, and link the resulting evidence from
`assets/`.

Suggested phase report names:

- `01-foundation.md`
- `02-session-api.md`
- `03-event-processing.md`
- `04-idempotency.md`
- `05-failure-recovery.md`

The expected eventual evidence set is:

1. `01-localstack-running`
2. `02-stack-resources`
3. `03-session-accepted`
4. `04-event-queued`
5. `05-dynamodb-items`
6. `06-idempotency-demo`
7. `07-consumer-logs`
8. `08-queue-buffering`
9. `09-recovery`
10. `10-dlq`

These names describe planned evidence only. No fake screenshots or unverified
results should be added. Reports and images must never expose LocalStack auth
tokens, AWS credentials, `.env` contents, or other secrets.
