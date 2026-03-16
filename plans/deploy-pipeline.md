---
name: deploy-pipeline
goal: Build a CI/CD pipeline that takes code from PR merge through staging validation to production deployment with automated rollback
diagram: |
    %%{init: {'theme': 'dark'}}%%
    graph TD
        A[PR Merged] --> B[Build & Test]
        B --> C{Tests Pass?}
        C -->|Yes| D[Deploy to Staging]
        C -->|No| E[Notify & Block]
        D --> F[Run Smoke Tests]
        F --> G{Healthy?}
        G -->|Yes| H[Deploy to Production]
        G -->|No| I[Rollback Staging]
        H --> J[Monitor 15min]
        J --> K{Metrics OK?}
        K -->|Yes| L[✓ Done]
        K -->|No| M[Auto-Rollback Prod]
status: active
session_count: 2
created: "2026-03-15"
updated: "2026-03-16"
current_task: 2
constraints:
    - Zero-downtime deployments only
    - All environments must use the same Docker image
    - Rollback must complete within 60 seconds
files:
    - path: .github/workflows/deploy.yml
      role: main pipeline definition
    - path: scripts/rollback.sh
      role: rollback automation
    - path: k8s/staging/deployment.yaml
      role: staging k8s config
    - path: k8s/prod/deployment.yaml
      role: production k8s config
tasks:
    - text: Set up Docker multi-stage build
      status: done
      spec: Create Dockerfile with build, test, and runtime stages
      verify:
        - ✓ docker build completes
        - ✓ image size under 200MB
      files:
        - Dockerfile
    - text: Configure GitHub Actions workflow
      status: done
      spec: Trigger on merge to main, run tests, build image, push to registry
      verify:
        - ✓ workflow triggers on merge
        - ✓ tests run in CI
      files:
        - .github/workflows/deploy.yml
    - text: Add staging deployment with smoke tests
      status: active
      spec: Deploy to staging namespace, run health checks and smoke test suite
      verify:
        - k8s deployment rolls out
        - smoke tests pass against staging URL
        - metrics dashboard shows healthy
      files:
        - k8s/staging/deployment.yaml
        - scripts/smoke-test.sh
    - text: Implement production deployment with canary
      status: todo
      spec: Canary deploy to 10% traffic, monitor, then full rollout
      verify:
        - canary receives 10% traffic
        - auto-promote after 15min if healthy
        - auto-rollback if error rate exceeds 1%
      files:
        - k8s/prod/deployment.yaml
    - text: Build automated rollback system
      status: todo
      spec: Watch error rate and latency, trigger rollback if thresholds exceeded
      verify:
        - rollback completes under 60s
        - alerts fire on rollback
      files:
        - scripts/rollback.sh
context:
    current_file: k8s/staging/deployment.yaml
    last_error: null
    test_state: 12 passing
    open_questions: Should we use Argo Rollouts or custom canary logic?
    pending_refactor: null
decisions:
    - time: 2026-03-15T09:00:00Z
      text: Use GitHub Actions over Jenkins — simpler, already integrated
    - time: 2026-03-15T14:00:00Z
      text: Docker multi-stage build to keep image small
    - time: 2026-03-16T10:00:00Z
      text: Staging smoke tests before any prod deployment
---

<!-- generated below — do not edit, use trail commands -->

## goal

Build a CI/CD pipeline that takes code from PR merge through staging validation to production deployment with automated rollback

## diagram

```mermaid
%%{init: {'theme': 'dark'}}%%
graph TD
    A[PR Merged] --> B[Build & Test]
    B --> C{Tests Pass?}
    C -->|Yes| D[Deploy to Staging]
    C -->|No| E[Notify & Block]
    D --> F[Run Smoke Tests]
    F --> G{Healthy?}
    G -->|Yes| H[Deploy to Production]
    G -->|No| I[Rollback Staging]
    H --> J[Monitor 15min]
    J --> K{Metrics OK?}
    K -->|Yes| L[✓ Done]
    K -->|No| M[Auto-Rollback Prod]
```

## constraints

- Zero-downtime deployments only
- All environments must use the same Docker image
- Rollback must complete within 60 seconds

## files

| path | role |
|---|---|
| .github/workflows/deploy.yml | main pipeline definition |
| scripts/rollback.sh | rollback automation |
| k8s/staging/deployment.yaml | staging k8s config |
| k8s/prod/deployment.yaml | production k8s config |

## tasks

- [x] 00 · Set up Docker multi-stage build
- [x] 01 · Configure GitHub Actions workflow
- [▶] 02 · Add staging deployment with smoke tests

  **spec:** Deploy to staging namespace, run health checks and smoke test suite

  **verify:**
  - [ ] k8s deployment rolls out
  - [ ] smoke tests pass against staging URL
  - [ ] metrics dashboard shows healthy

  **files:** k8s/staging/deployment.yaml, scripts/smoke-test.sh

- [ ] 03 · Implement production deployment with canary
- [ ] 04 · Build automated rollback system

## context

| field | value |
|---|---|
| current_file | k8s/staging/deployment.yaml |
| last_error | ~ |
| test_state | 12 passing |
| open_questions | Should we use Argo Rollouts or custom canary logic? |
| pending_refactor | ~ |

## decisions

- 2026-03-15 · Use GitHub Actions over Jenkins — simpler, already integrated
- 2026-03-15 · Docker multi-stage build to keep image small
- 2026-03-16 · Staging smoke tests before any prod deployment

## notes
