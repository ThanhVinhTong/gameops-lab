# Runtime evidence assets

Only real runtime evidence belongs in this directory.

- Prefer terminal output that demonstrates behavior.
- Do not use screenshots of source code as runtime evidence.
- Never expose LocalStack authentication tokens.
- Never expose AWS credentials.
- Never expose `.env` or its contents.
- Dynamic IDs and ARNs normally do not need redaction unless they contain
  sensitive information.
- Remove unrelated sensitive information and image metadata before committing.
- Name images consistently and reference each image from a report.

Suggested names:

- `01-localstack-running.png`
- `02-stack-resources.png`
- `03-session-accepted.png`
- `04-event-queued.png`
- `05-dynamodb-items.png`
- `06-idempotency-demo.png`
- `07-consumer-logs.png`
- `08-queue-buffering.png`
- `09-recovery.png`
- `10-dlq.png`

When a completed report lives directly in `reports/`, reference an image as
`assets/<filename>`.
