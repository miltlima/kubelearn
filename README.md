# KubeLearn Project Documentation

## Overview

KubeLearn is a tool designed to help users test and expand their knowledge of Kubernetes. This project includes both a backend, which runs various Kubernetes-related tasks, and a frontend, which provides a user interface for interacting with the tool.

## Project Structure

The project is organized into the following directories:

```bash
.
├── cmd
│   └── main.go                # Backend Go application
├── kubelearn-frontend         # Frontend React application
│   ├── postcss.config.js
│   ├── src
│   │   ├── App.js
│   │   ├── App.test.js
│   │   ├── index.js
│   │   ├── reportWebVitals.js
│   │   └── setupTests.js
│   └── tailwind.config.js
├── makefile                   # Makefile for managing the project
└── pkg
    ├── k8s                    # Kubernetes-related utilities
    │   └── client.go
    ├── resources              # Contains Kubernetes-related questions
    │   ├── easy
    │   ├── hard
    │   └── medium
    └── utils                  # Additional utilities
```
## Prerequisites

To run the project locally, you need to have the following installed:

- Go (Golang)
- Node.js and npm
- Terraform
- Kubernetes (kubectl)
- Kind (for local Kubernetes clusters)

## Makefile Commands

The `Makefile` provides several commands to manage the project efficiently. Below is a summary of the available targets:

### Setup and Run KubeLearn

This command initializes Terraform, sets up the backend and frontend, and starts the services.

```sh
make Kubelearn
```

### Stop KubeLearn

This command stops the backend and frontend services.

```sh
make stopKubelearn
```

### Terraform Commands

- **init**: Initializes the Terraform configuration.

  ```sh
  make init
  ```

- **apply**: Applies the Terraform configuration.

  ```sh
  make apply
  ```

- **destroy**: Destroys the Terraform-managed infrastructure.

  ```sh
  make destroy
  ```

### Kubernetes Manifest Management

- **all**: Installs all Kubernetes manifests from the `manifests` directory.

  ```sh
  make all
  ```

- **clean**: Deletes all installed Kubernetes resources.

  ```sh
  make clean
  ```

- **check-syntax**: Validates the syntax of all Kubernetes manifests without applying them.

  ```sh
  make check-syntax
  ```

## Running the Project

1. **Setup Environment**: Run `make Kubelearn` to initialize Terraform and start both the backend and frontend.

2. **Access the Application**: After running the setup, the frontend will be available at `http://localhost:3000` by default, and the backend API at `http://localhost:8083`.

3. **Stop the Services**: Run `make stopKubelearn` to stop the backend and frontend services.

## Additional Information

- Logs for the backend and frontend are stored in `backend.log` and `frontend.log`, respectively.
- Ensure that your Kubernetes environment is properly configured before running the application.

## Troubleshooting

- If Terraform fails to initialize, ensure that your Terraform installation is correct and that the `config` directory contains valid configurations.
- If the frontend doesn't start, verify that all npm dependencies are installed correctly.

---

This documentation should cover the basic setup and usage of the KubeLearn project. For more detailed instructions or troubleshooting, please refer to the specific sections in this document.
``` 

You can use this documentation as a starting point and expand it further as needed. It covers the basic commands, project structure, and steps to run the project locally.
```

## Answer the questions below

| Questions | Description                                                                                                                                                            |
|-----------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 1         | Create a new pod called `nginx` with `nginx:alpine` image in `default` namespace.                                                                                      |
| 2         | Create a new deployment called `nginx-deployment` with `nginx:alpine` image and `4 replicas` in default namespace.                                                     |
| 3         | Create a new deployment called `redis` with image `redis:alpine` in `latam` namespace, and create a service called `redis-service` with port `6379` in same namespace. |
| 4         | Create a namespace called `europe`                                                                                                                                     |
| 5         | Create a configmap `europe-configmap` with data `France=Paris`                                                                                                         |
| 6         | Create a pod `thsoot` with label `country=china`, `amazon/amazon-ecs-network-sidecar:latest` image and namespace `asia`                                                |
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


