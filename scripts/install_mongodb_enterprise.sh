#!/usr/bin/env bash

set -e

NAMESPACE=mongodb
echo "Installing MongoDB Kubernetes Operator..."
mkdir /database/enterprises/ && cd /database/enterprises/
git clone https://github.com/mongodb/mongodb-enterprise-kubernetes.git
cd ./mongodb-enterprise-kubernetes/
kubectl create namespace ${NAMESPACE}
helm install mongodb-enterprise helm_chart --values helm_chart/values.yaml --set namespace=${NAMESPACE}

echo "Configuring kubectl to default namespace..."
kubectl config set-context $(kubectl config current-context) --namespace=${NAMESPACE}

echo "Creating ConfigMap..."
kubectl create configmap mongoconfigmap \
    --from-literal="baseUrl=" \
    --from-literal="projectName="