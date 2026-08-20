# AWS adapters

This directory is reserved for the AWS-specific adapters that will be added in
later phases. Those adapters are expected to cover:

- EventBridge event publishing
- DynamoDB analytics persistence
- AWS SDK client construction when it becomes necessary

> No AWS client implementation exists in the initial repository foundation.

Application and domain logic stays in `internal/gameops`; environment loading
stays in `internal/config`.
