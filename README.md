<p align="center">
  <img src="./assets/kubelearn_logo.png" alt="Kubelearn Logo" width="250"/>
</p>

<h1 align="center">Kubelearn</h1>

<p align="center">
  Test your knowledge of Kubernetes — interactive quizzes, timed challenges, and score tracking.
</p>

## Overview

KubeLearn helps you test and expand your Kubernetes knowledge. It provisions a local [kind](https://kind.sigs.k8s.io/) cluster, gives you 26 hands-on tasks, and checks your cluster to tell you which ones you solved.

It has two parts:

- **Backend** (Go): inspects the cluster and reports which questions pass. It also serves the in-browser terminal.
- **Frontend** (React): shows the questions, a timer, your score, and a terminal to run `kubectl`.

## Project Structure

```bash
.
├── cmd                        # Backend entry point (HTTP API and terminal)
├── pkg
│   ├── k8s                    # Kubernetes client setup
│   ├── resources              # Question checkers, grouped by difficulty
│   │   ├── easy
│   │   ├── medium
│   │   └── hard
│   └── utils                  # Shared helpers
├── kubelearn-frontend         # React application
├── config                     # Terraform configuration for the kind cluster
├── manifests                  # Lab resources applied to the cluster
├── manifests_examples         # Example answers
└── Makefile                   # Automation for the whole project
```

## Prerequisites

- Go 1.21 or newer
- Node.js and npm
- Terraform
- kubectl
- kind (and Docker, which kind needs)
- make

## Quick Start

1. Provision the cluster and start the backend and frontend:

   ```sh
   make kubelearn
   ```

2. Apply the lab resources that some questions depend on:

   ```sh
   make manifests-apply
   ```

3. Open the frontend at `http://localhost:3000`. The backend API listens on `http://127.0.0.1:8083`.

4. When you are done, stop the services:

   ```sh
   make stop-kubelearn
   ```

   To delete the cluster as well, run `make terraform-destroy`.

## Makefile Commands

Run `make help` to list every target and variable.

| Target | What it does |
|--------|--------------|
| `kubelearn` | Provision the cluster, switch the kubectl context, start backend and frontend |
| `stop-kubelearn` | Stop backend and frontend |
| `clean` | Stop the services and remove the backend binary and logs |
| `infra-up` | Provision the cluster and switch the kubectl context |
| `terraform-init` / `terraform-plan` / `terraform-apply` / `terraform-destroy` | Manage the kind cluster with Terraform |
| `kube-context` | Switch kubectl to the KubeLearn context |
| `manifests-apply` | Apply every manifest in `manifests/` |
| `manifests-delete` | Delete the resources from `manifests/` |
| `manifests-lint` | Validate the manifests with a client-side dry run |
| `backend-build` / `backend-start` / `backend-stop` | Build, start, or stop the backend |
| `frontend-install` / `frontend-start` / `frontend-stop` | Install dependencies, start, or stop the frontend |

Variables you can override, for example `make kubelearn LOG_DIR=/tmp/kubelearn`:

| Variable | Default | Purpose |
|----------|---------|---------|
| `KUBE_CONTEXT` | `kind-kubelearn` | kubectl context to switch to |
| `MANIFESTS_DIR` | `manifests` | Directory with the lab manifests |
| `TERRAFORM_DIR` | `config` | Terraform configuration directory |
| `FRONTEND_DIR` | `kubelearn-frontend` | Frontend directory |
| `BACKEND_BIN` | `bin/kubelearn` | Where the backend binary is built |
| `LOG_DIR` | `logs` | Runtime logs and PID files |

## Backend

The backend finds the cluster the same way `kubectl` does: the files in the `KUBECONFIG` environment variable, or `~/.kube/config`, using the current context. The cluster must exist before the backend starts.

`/questions` and `/finish` give the cluster 20 seconds to answer all checks; after that they return `504` instead of hanging.

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/questions` | GET | List the questions and whether each one currently passes |
| `/finish` | GET | Return the score as a percentage |
| `/start` | GET | Mark the quiz as started |
| `/setup` | POST | Run `make infra-up` and `make manifests-apply` in the background |
| `/terminal` | WebSocket | Run shell commands typed in the frontend terminal |

Environment variables:

| Variable | Default | Purpose |
|----------|---------|---------|
| `KUBELEARN_ALLOWED_ORIGINS` | `http://localhost:3000,http://127.0.0.1:3000` | Comma-separated browser origins allowed to call the API and open the terminal |
| `KUBELEARN_ROOT` | working directory | Directory with the project Makefile, used by `/setup` |

### Security

The terminal endpoint runs any shell command it receives, with your user's permissions. For that reason the backend listens on `127.0.0.1` only and rejects requests from origins that are not in `KUBELEARN_ALLOWED_ORIGINS`. Do not expose port 8083 to a network.

## Development

```sh
go test ./...                                              # backend tests
go vet ./...
npm --prefix kubelearn-frontend test -- --watchAll=false   # frontend tests
```

The questions live in a single registry, `pkg/resources/registry.go`. To add one, write its checker under `pkg/resources/<difficulty>/` and append it to `Checks`.

Every push to `main` and every pull request runs the build, formatting checks, and tests for both parts (see `.github/workflows/ci.yaml`). Releases are built by GoReleaser when a tag is pushed (see `.github/workflows/release.yaml`).

## Troubleshooting

- **Logs**: backend and frontend output go to `logs/backend.log` and `logs/frontend.log`.
- **Backend exits at startup**: it could not load a kubeconfig (`KUBECONFIG` or `~/.kube/config`). Run `make infra-up` first.
- **Questions fail to load with a timeout**: the cluster is not answering. Check `kubectl get nodes`.
- **Terminal shows "Connection error"**: check that the backend is running and that the frontend origin is in `KUBELEARN_ALLOWED_ORIGINS` (the dev server may pick a port other than 3000 if it is busy).
- **Context `kind-kubelearn` not found**: the cluster does not exist. Run `make infra-up`.
- **Terraform fails to initialize**: check your Terraform installation and that Docker is running.
- **Frontend does not start**: run `make frontend-install` and check `logs/frontend.log`.

## Answer the questions below

| Questions | Description                                                                                                                                                            |
|-----------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 1         | Create a new pod called `nginx` with `nginx:alpine` image in `default` namespace.                                                                                      |
| 2         | Create a new deployment called `nginx-deployment` with `nginx:alpine` image and `4 replicas` in default namespace.                                                     |
| 3         | Create a new deployment called `redis` with image `redis:alpine` in `latam` namespace, and create a service called `redis-service` with port `6379` in same namespace. |
| 4         | Create a namespace called `europe`                                                                                                                                     |
| 5         | Create a configmap `europe-configmap` with data `France=Paris`                                                                                                         |
| 6         | Create a pod `tshoot` with label `country=china`, `amazon/amazon-ecs-network-sidecar:latest` image and namespace `asia`                                                |
| 7         | Create a persistent volume `unicorn-pv` with capacity `1Gi` and access mode `ReadWriteMany` and host path `/tmp/data`                                                  |
| 8         | Create a persistent volume claim `unicorn-pvc` with capacity `400Mi` and access mode `ReadWriteMany`                                                                   |
| 9         | Create a pod `webserver` in `public` namespace with `nginx:alpine` image and a volume mount `/usr/share/nginx/html` and a persistent volume claim `unicorn-pvc`        |
| 10        | There is a pod with problem, Can you able to solve it ?                                                                                                                |
| 11        | Create a network policy `allow-policy-colors` to allow `redmobile-webserver` to access `bluemobile-dbcache` (There objects are created in colors namespace)            |
| 12        | Create a secret `secret-colors` with data `color=red` in `colors` namespace                                                                                            |
| 13        | Add a secret `secret-purple` with data `singer=prince` to the pod `purple` with image `redis:alpine` in `colors` namespace                                             |
| 14        | Create a service account `america-sa` in `default` namespace                                                                                                           |
| 15        | Add service account `america-sa` to the deployment `mark42`                                                                                                            |
| 16        | Change the replica count of the deployment `mark42` to `5`                                                                                                             |
| 17        | Create a horizontal pod autoscaler for deployment `mark43` with cpu percent `80`, min replicas `2` and max replicas `8`                                                |
| 18        | Prevent privilege escalation in the deployment `mark42`                                                                                                                |
| 19        | Add a `liveness` probe to the pod `mark50` with initial delay `5s`, period `10s` and path `/` in namespace `shield`                                                    |
| 20        | Create a deployment `yellow-deployment` with `bonovoo/node-app:1.0` image and `2` replicas in namespace `colors`                                                       |
| 21        | Create a service `yellow-service` for the deployment `yellow-deployment` in namespace `colors` with port `80` and target port `3000`                                   |
| 22        | Create an ingress `ingress-colors` with host `yellow.com`, path `/yellow` and service `yellow-service` in namespace `colors`                                           |
| 23        | Create a role `apple-one` with verbs `get, list, watch` in namespace `fruits`                                                                                          |
| 24        | Create a job `job-gain` with parallelism `2`, completions `4`, backoffLimit `3` and deadlineSeconds `40`                                                               |
| 25        | Create a cronjob `cronjob-gain` run a each `5` minutes with image `busybox:1.28`, command `'sleep 3600'` and restartPolicy `Never`                                     |
| 26        | Create a statefulset `statefulset-gain` with image `busybox:1.28`, command `'sleep 3600'` and replicas `3`                                                             |


