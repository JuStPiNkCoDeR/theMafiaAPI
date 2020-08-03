#!/usr/bin/env bash

set -e

#echo "Checking of dependencies..."
#sudo ./install_dependencies.sh
#echo "Checking of dependencies successfully complete."

echo "Starting the Minikube..."
minikube start --driver=kvm2

# For using local docker images
eval $(minikube docker-env)

echo "Deploying services..."
echo ""
for servicePath in ./services/*/; do
    echo "Moving to $servicePath directory..."
    cd "$servicePath"
    serviceName=${PWD##*/}
    echo "Building the $serviceName docker image..."
    docker build -t ${serviceName} .
    echo "Applying Kubernetes configuration..."
    kubectl apply -f k8s-deployment.yml
    echo ""
done

echo "Successfully deployed."
