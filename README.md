# Kubelearn - Learn how to deploy objects in kubernetes

## how to use

clone the repository

```bash
git clone git@github.com:miltlima/kubelearn.git
```

Enter inside folder kubelearn

```bash
cd kubelearn
```

Runnning the following command:

```bash
go run main.go
```

This will show the table above:

```bash
+---------------------------------------------------------------------------------------------------------------------+--------+------------+
|                                                KUBELEARN - TEST YOUR                                                | RESULT | DIFFICULTY |
|                                               KNOWLEDGE OF KUBERNETES                                               |        |            |
+---------------------------------------------------------------------------------------------------------------------+--------+------------+
| Question 1 - Deploy a POD nginx name with nginx:alpine image                                                        | ✅Pass | Easy       |
| Question 2 - Create a deployment nginx-deployment with nginx:alpine image and 4 replicas                            | 🆘Fail | Medium     |
| Question 3 - Create a deployment redis name with redis:alpine image and a service with port 6379 in namespace latam | 🆘Fail | Hard       |
+---------------------------------------------------------------------------------------------------------------------+--------+------------+
```

## Questions

1 - Create a new pod called `nginx` with `nginx:latest` image in default namespace.

2 - Create a new deployment called `nginx-deployment` with `nginx:1.17` image and `4 replicas` in default namespace.

3 - Create a new deployment called `redis` with image `redis:alpine` in `latam` namespace , and create a service called `redis-service` with port `6379` in same namespace.
