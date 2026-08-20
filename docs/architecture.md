# GameOps Lab Architecture

## TARGET ARCHITECTURE

This diagram describes the intended end state. The AWS resources and application
flow are not implemented in the repository foundation.

```mermaid
flowchart TD
    Client[Client] --> API[API Gateway]
    API --> Session[Session API Lambda]
    Session --> Events[EventBridge]
    Events --> Queue[SQS Analytics Queue]
    Queue --> Consumer[Analytics Consumer Lambda]
    Queue -. failed messages .-> DLQ[SQS DLQ]
    Consumer --> Database[DynamoDB]
```

| Component | Local implementation | AWS production equivalent |
| --- | --- | --- |
| API | future LocalStack/SAM | API Gateway |
| compute | local Lambda emulation | AWS Lambda |
| event routing | LocalStack EventBridge | Amazon EventBridge |
| queue | LocalStack SQS | Amazon SQS |
| database | LocalStack DynamoDB | Amazon DynamoDB |
| infrastructure | SAM + LocalStack CloudFormation | AWS CloudFormation |

All rows in this table are planned; none have been runtime verified.

## SAM and CloudFormation

**AWS CloudFormation** is the Infrastructure-as-Code engine. **AWS SAM** is a
serverless extension of CloudFormation. Both SAM and native CloudFormation
resources will live in
[`infra/cloudformation/template.yaml`](../infra/cloudformation/template.yaml),
the single GameOps infrastructure definition.

| Example type | Classification |
| --- | --- |
| `AWS::Serverless::Function` | SAM resource |
| `AWS::Serverless::HttpApi` | SAM resource |
| `AWS::SQS::Queue` | Native CloudFormation resource |
| `AWS::DynamoDB::Table` | Native CloudFormation resource |
| `AWS::Events::EventBus` | Native CloudFormation resource |

SAM resources are deployed through CloudFormation: SAM transforms its
higher-level serverless types into CloudFormation-compatible resources.

In the future integration stage, `sam build` will build the serverless project
and `samlocal deploy` will deploy the SAM/CloudFormation stack to LocalStack.
The real-AWS command `sam deploy` must not be used for this local-only project
unless a later decision explicitly enables real AWS.

## Delivery sequence

1. Repository foundation
2. Go domain and application code
3. Unit tests
4. AWS adapters
5. SAM / CloudFormation resources
6. LocalStack execution and integration verification
7. Reports and screenshots

AWS-related commands begin only at step 6. This ordering keeps infrastructure
execution separate from application design and unit-level verification.
