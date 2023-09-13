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

![Kubelearn](images/kubelearn.png)

## Questions

1 - Create a new pod called `nginx` with `nginx:alpine` image in default namespace.

2 - Create a new deployment called `nginx-deployment` with `nginx:1.17` image and `4 replicas` in default namespace.

3 - Create a new deployment called `redis` with image `redis:alpine` in `latam` namespace , and create a service called `redis-service` with port `6379` in same namespace.

4 - Create a namespace called `emea`

5 - Create a configmap emea-configmap with data France=Paris
