# Kubelearn

Practice you kubernetes knowledge

Requirements:

- Golang 1.21
- Terraform 1.5

## how to use

**Clone the repository**

```bash
git clone git@github.com:miltlima/kubelearn.git
```

**Enter inside folder kubelearn**

```bash
cd kubelearn
```

## Optionally you can install kind using make ( requires Docker)

**Initialize Terraform repository**:

```sh
make init
```

**Apply Terraform configurations**:

```sh
make apply
```

**Destroy Terraform resources**

```sh
make destroy
```

## You can also use `make help` for list available commands

- `all`: Installs all YAML manifests.
- `clean`: Deletes all installed resources.
- `check-syntax`: Checks the syntax of all manifests without actually installing them.
- `init`: Initializes the Terraform repository.
- `apply`: Applies Terraform configurations.
- `destroy`: Destroys Terraform resources.

## Prepare environment ( this will create kubernetes objects needed for some questions)

**Install YAML manifests**:

```sh
make all
```

**Clean up installed resources**:

```sh
make clean
```

**Check manifest syntax**:

```sh
make check-syntax
```

## Answer the questions below

| Questions   | Description |
| ----------- | ----------- |
| 1 | Create a new pod called `nginx` with `nginx:alpine` image in `default` namespace.|
| 2 | Create a new deployment called `nginx-deployment` with `nginx:alpine` image and `4 replicas` in default namespace.|
| 3 | Create a new deployment called `redis` with image `redis:alpine` in `latam` namespace, and create a service called `redis-service` with port `6379` in same namespace.|
| 4 | Create a namespace called `europe`|
| 5 | Create a configmap `europe-configmap` with data `France=Paris`|
| 6 | Create a pod `thsoot` with label `country=china`, `amazon/amazon-ecs-network-sidecar:latest` image and namespace `asia`|
| 7 | Create a persistent volume `unicorn-pv` with capacity `1Gi` and access mode `ReadWriteMany` and host path `/tmp/data`|
| 8 | Create a persistent volume claim `unicorn-pvc` with capacity `400Mi` and access mode `ReadWriteMany`|
| 9 | Create a pod `webserver` in `public` namespace with `nginx:alpine` image and a volume mount `/usr/share/nginx/html` and a persistent volume claim `unicorn-pvc`|
| 10| There is a pod with problem, Can you able to solve it ?|
| 11| Create a network policy `allow-policy-colors` to allow `redmobile-webserver` to access `bluemobile-dbcache` (There objects are created in colors namespace)|
| 12| Create a secret `secret-colors` with data `color=red` in `colors` namespace|
| 13| Add a secret `secret-purple` with data `singer=prince` to the pod `purple` with image `redis:alpine` in `colors` namespace|
| 14| Create a service account `america-sa` in `default` namespace|
| 15| Add service account `america-sa` to the deployment `mark42`|
| 16| Change the replica count of the deployment `mark42` to `5`|
| 17| Create a horizontal pod autoscaler for deployment `mark43` with cpu percent `80`, min replicas `2` and max replicas `8`|
| 18| Prevent privilege escalation in the deployment `mark42`|
| 19| Add a `liveness` probe to the pod `mark50` with initial delay `5s`, period `10s` and path `/` in namespace `shield`|
| 20| Create a deployment `yellow-deployment` with `bonovoo/node-app:1.0` image and `2` replicas in namespace `colors`|
| 21| Create a service `yellow-service` for the deployment `yellow-deployment` in namespace `colors` with port `80` and target port `3000`|
| 22| Create an ingress `ingress-colors` with host `yellow.com`, path `/yellow` and service `yellow-service` in namespace `colors`|
| 23| Create a role `role-one` with verbs `get, list, watch` in namespace `default`|  

## Running the following command

```bash
go run main.go
```

## This will show the table above

![Kubelearn](images/kubelearn.png)
