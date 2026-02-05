# Todo Service (Go)

A simple, testable TODO microservice with repository pattern, Postgres persistence, and concurrent bulk-create using a worker pool. This service currently exposes REST JSON endpoints for convenience; gRPC contracts are defined in protobuf for service-to-service communication.

## Highlights
- Repository pattern (swap Postgres for MongoDB later without changing service logic).
- Bulk create uses a 4-worker pool and is all-or-nothing.
- Clean separation: transport -> service -> repository.
- Dev infra with Docker + K8s + Tilt.
- Production infra with EKS-ready manifests + Dockerfile.

## Project Structure
- `cmd/todo-api/main.go` - entrypoint.
- `cmd/migrate/main.go` - migration CLI.
- `internal/config` - config from env.
- `internal/model` - domain models.
- `internal/repository` - interface + Postgres + in-memory repo.
- `internal/service` - business logic + validation.
- `internal/worker` - generic worker pool.
- `internal/transport/http` - REST handlers.
- `internal/transport/grpc` - gRPC server implementation.
- `proto` - protobuf contract.
- `shared/gen` - generated protobuf stubs shared across services.
- `migrations` - SQL schema.
- `infra/dev` - dev Docker + K8s + Tilt.
- `infra/production` - production Docker + K8s for EKS.

## Configuration
Environment variables:
- `TODO_HTTP_ADDR` (default `:8080`)
- `TODO_GRPC_ADDR` (default `:9090`)
- `TODO_DB_DSN` (default `postgres://todo:todo@localhost:5432/todo?sslmode=disable`)
- `TODO_WORKERS` (default `4`)

## REST API
Base URL: `http://localhost:8080`

- `POST /todos`
  - body: `{ "title": "...", "description": "..." }`
- `GET /todos?limit=50&offset=0`
- `GET /todos/{id}`
- `PATCH /todos/{id}`
  - body: `{ "title": "...", "description": "...", "status": "pending|done" }`
- `DELETE /todos/{id}`
- `POST /todos/bulk`
  - body: `{ "items": [{"title":"...","description":"..."}, ...] }`
  - all-or-nothing: any validation error rejects the entire batch.

## Protobuf / gRPC
Proto definition: `proto/todo/v1/todo.proto`

Generate Go code (requires `protoc` + protoc-gen-go + protoc-gen-go-grpc):
```
protoc --proto_path=proto/todo/v1 \
  --go_out=shared/gen/todo/v1 --go-grpc_out=shared/gen/todo/v1 \
  --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative \
  proto/todo/v1/todo.proto
```

gRPC server starts on `TODO_GRPC_ADDR` (default `:9090`).

## Migrations
Run migrations using the built-in CLI:

```
go run ./cmd/migrate -database "postgres://todo:todo@localhost:5432/todo?sslmode=disable" -command up
```

Other commands:
- `-command down`
- `-command steps -steps 1`
- `-command goto -version 1`
- `-command version`

## Convenience Scripts
If you don’t have `make` on Windows, use PowerShell scripts:

```
./scripts/generate.ps1
./scripts/migrate.ps1 -Dsn "postgres://todo:todo@localhost:5432/todo?sslmode=disable" -Command up
./scripts/run.ps1
```

### Installing Make on Windows (Optional)
- **Scoop**: `scoop install make`
- **Chocolatey**: `choco install make`
- **MSYS2**: `pacman -S make`

Then you can run:
```
make generate-proto
make migrate-up DB_DSN="postgres://todo:todo@localhost:5432/todo?sslmode=disable"
make run
```

## Development With Tilt
Prereqs: Docker Desktop, kubectl, Tilt.

```
cd infra/dev

tilt up
```

This builds the dev Docker image, applies K8s manifests, and forwards port `8080`.

## Production (AWS EKS)
Below is a detailed step-by-step AWS deployment guide. This uses **EKS**, **ECR**, **IAM**, **VPC**, **EC2**, and **ELB**. Replace placeholders such as `YOUR_AWS_ACCOUNT_ID` and `YOUR_REGION`.

### 1) Install Tooling
- AWS CLI
- kubectl
- eksctl
- Docker

### 2) Configure AWS CLI
```
aws configure
```

### 3) IAM Setup (Minimal Guidance)
You can do this via the AWS Console or CLI. Common roles/policies:

- **Cluster role** (for EKS control plane):
  - `AmazonEKSClusterPolicy`
  - `AmazonEKSServicePolicy`

- **Node group role** (for worker nodes):
  - `AmazonEKSWorkerNodePolicy`
  - `AmazonEKS_CNI_Policy`
  - `AmazonEC2ContainerRegistryReadOnly`

- **ECR Push Access** (for your user or CI):
  - `AmazonEC2ContainerRegistryFullAccess` (or a scoped policy to just your repo)

If you need a fast path for early development, you can attach `AdministratorAccess` to a dedicated dev user, but replace it with least-privilege as soon as possible.

### 4) Create an ECR Repository
```
aws ecr create-repository --repository-name todo-api --region YOUR_REGION
```

### 5) Login to ECR
```
aws ecr get-login-password --region YOUR_REGION | \
  docker login --username AWS --password-stdin YOUR_AWS_ACCOUNT_ID.dkr.ecr.YOUR_REGION.amazonaws.com
```

### 6) Build and Push the Production Image
```
docker build -f infra/production/Dockerfile -t todo-api:latest .

docker tag todo-api:latest YOUR_AWS_ACCOUNT_ID.dkr.ecr.YOUR_REGION.amazonaws.com/todo-api:latest

docker push YOUR_AWS_ACCOUNT_ID.dkr.ecr.YOUR_REGION.amazonaws.com/todo-api:latest
```

### 7) Create an EKS Cluster
Example using `eksctl`:
```
eksctl create cluster \
  --name todo-prod \
  --region YOUR_REGION \
  --nodes 2 \
  --nodegroup-name todo-nodes \
  --node-type t3.medium
```

This creates the control plane, VPC networking, and a managed node group.

### 8) Update kubeconfig
```
aws eks update-kubeconfig --name todo-prod --region YOUR_REGION
```

### 9) Apply Kubernetes Manifests
Update the image in `infra/production/k8s/todo-api.yaml` with your ECR image.

```
kubectl apply -f infra/production/k8s/namespace.yaml
kubectl apply -f infra/production/k8s/configmap.yaml
kubectl apply -f infra/production/k8s/secret.yaml
kubectl apply -f infra/production/k8s/postgres.yaml
kubectl apply -f infra/production/k8s/todo-api.yaml
```

### 10) Get the Load Balancer URL
```
kubectl get svc -n todo-prod
```
Look for the external `EXTERNAL-IP` on the `todo-api` service.

### 11) Production Database (Recommended)
The included `infra/production/k8s/postgres.yaml` uses `emptyDir` (not durable). For real production:
- Use **RDS for PostgreSQL**.
- Update `infra/production/k8s/secret.yaml` with your RDS connection string.
- Remove the in-cluster `postgres.yaml` deployment.

## Commit Steps (Suggested)
1. Scaffold + proto + migration files.
2. Service + repository + worker pool + HTTP handlers + tests.
3. Dev infra (Docker/K8s/Tilt).
4. Production infra (Docker + EKS manifests).
5. README with AWS deployment steps.
6. gRPC server + shared stubs + tooling scripts.

## Notes
- The REST API is for local/dev. In production, an API Gateway can expose REST while service-to-service communication uses gRPC.
- You can swap Postgres for MongoDB by adding a new repository implementation under `internal/repository`.