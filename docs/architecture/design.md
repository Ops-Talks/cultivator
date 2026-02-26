# Architecture Design

Understand how Cultivator works under the hood.

## Overview

Cultivator consists of several key components that work together to provide automated Terragrunt workflows.

## Components

### 1. Event Handler
- Receives GitHub events (PR opened, comment created)
- Validates webhook signatures
- Routes to appropriate handlers

### 2. Change Detector
- Analyzes Git diffs between base and head branches
- Identifies affected Terragrunt modules
- Builds list of modules to process

### 3. Dependency Graph
- Parses Terragrunt configurations
- Builds dependency graph between modules
- Determines execution order

### 4. Executor
- Runs Terragrunt commands (plan, apply)
- Manages locks to prevent concurrent operations
- Captures output and errors

### 5. Formatter
- Formats Terragrunt output
- Redacts sensitive information
- Creates PR comments with results

### 6. GitHub Integration
- Posts comments on PRs
- Updates commit statuses
- Manages PR reviews

## Data Flow

```
GitHub Event
    ↓
Event Handler
    ↓
Change Detector → Git Analysis
    ↓
Dependency Graph → Terragrunt Config Parser
    ↓
Executor → Terragrunt Commands
    ↓
Formatter → Output Processing
    ↓
GitHub Integration → PR Comment
```

## Key Features Implementation

### Smart Change Detection
- Analyzes file changes in the PR
- Determines which modules are affected
- Respects Terragrunt include/exclude patterns

### Dependency-Aware Execution
- Parses `dependency` blocks in `terragrunt.hcl`
- Builds execution graph
- Runs modules in correct order

### Locking Mechanism
- Uses file-based locks in shared storage
- Prevents concurrent applies
- Configurable timeout and retry

### Security
- Redacts sensitive data from outputs
- Validates GitHub webhook signatures
- Respects IAM permissions

## Configuration Storage

Cultivator reads from `cultivator.yml` in the repository root:

```yaml
version: 1
settings:
  auto_plan: true
  require_approval: false
  lock_timeout: 30m
```

See [Configuration](../getting-started/configuration.md) for full reference.
