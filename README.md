# GameOps Lab

A local-first, AWS-compatible event-driven gaming analytics platform built with Go.

## Overview

GameOps Lab captures gaming sessions and processes them asynchronously to generate weekly playtime analytics. The project demonstrates how an event-driven workload can be developed locally while retaining a documented migration path to production AWS services.

## Architecture

```text
Client
  |
  v
API Gateway + Session Lambda
  |
  v
EventBridge
  |
  v
SQS Queue ------> Dead-Letter Queue
  |
  v
Analytics Lambda
  |
  v
DynamoDB Local
```

## Technology Stack

* Go
* AWS SAM
* API Gateway
* AWS Lambda
* Amazon EventBridge
* Amazon SQS and dead-letter queues
* DynamoDB Local
* LocalStack
* Docker Compose

## Engineering Focus

* Event-driven processing
* Retry and dead-letter queue handling
* Idempotent consumers
* DynamoDB conditional writes
* Infrastructure as code
* Local-to-AWS portability
* Reliability, security, and cost trade-offs

## Project Status

Core event-processing flow under active implementation.

## Local and Production Mapping

| Local development      | Production target      |
| ---------------------- | ---------------------- |
| AWS SAM local API      | Amazon API Gateway     |
| Lambda containers      | AWS Lambda             |
| LocalStack EventBridge | Amazon EventBridge     |
| LocalStack SQS         | Amazon SQS             |
| DynamoDB Local         | Amazon DynamoDB        |
| Docker logs            | Amazon CloudWatch Logs |

## Planned Validation Scenarios

1. Submit gaming session events while the analytics consumer is unavailable.
2. Restore the consumer and verify queued messages are processed.
3. Submit the same event twice and verify analytics are updated only once.
4. Submit malformed events and verify they are moved to the dead-letter queue.

## Timeline

* July 2026: Initial concept, requirements analysis, and architecture design
* August 2026: Local implementation and integration testing
