#!/usr/bin/env bash

# The script installs all dependencies for Ubuntu/Debain.

set -e

echo "Begin to read dependencies.txt file at ${PWD}"
# Read dependencies text file and check them with versions if passed
while IFS= read -r line; do
     echo "Read line: $line"
     cmd=${line%=*}
     version=${line#*=}
     echo "Got command name: $cmd and it's version: $version"

     echo "Check if installed the $cmd"
     if ! command -v ${cmd} &> /dev/null
     then
        echo "The $cmd could't be found."
        echo "Installing..."

        # Non-simple ways
        if [[ "$cmd" = "kubectl" ]]; then
            echo "Detected specific installation instructions for Kubectl."
            apt-get update &&  apt-get install -y apt-transport-https gnupg2
            curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg |  apt-key add -
            echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" |  tee -a /etc/apt/sources.list.d/kubernetes.list
        elif [[ "$cmd" = "minikube" ]]; then
            echo "Detected specific installation instructions for Minikube."
            curl -Lo minikube https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64 \
                && chmod +x minikube
             mkdir -p /usr/local/bin/
             install minikube /usr/local/bin/
        elif [[ "$cmd" = "docker" ]]; then
            echo "Detected specific installation instructions for Docker."
            apt-get install \
                 apt-transport-https \
                 ca-certificates \
                 curl \
                 gnupg-agent \
                 software-properties-common
            curl -fsSL https://download.docker.com/linux/ubuntu/gpg |  apt-key add -
            apt-key fingerprint 0EBFCD88
            add-apt-repository \
                 "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
                 $(lsb_release -cs) \
                 stable"
            apt-get update
            apt-get install docker-ce docker-ce-cli containerd.io
        elif [[ "$cmd" = "helm" ]]; then
            echo "Detected specific installation instructions for Helm."
            curl https://helm.baltorepo.com/organization/signing.asc | apt-key add -
            apt-get install apt-transport-https --yes
            echo "deb https://baltocdn.com/helm/stable/debian/ all main" | tee /etc/apt/sources.list.d/helm-stable-debian.list
            apt-get update
            apt-get install helm

            echo "Add official Helm stable charts..."
            helm repo add stable https://kubernetes-charts.storage.googleapis.com/
        else
            apt-get update
            apt-get install -y ${cmd}
        fi

        echo "Successfully installed."
     elif [[ ! "$version" = "*" ]]
     then
        current_version="$(${cmd} --version)"
        if [[ "$(printf '%s\n' "$version" "$current_version" | sort -V | head -n1)" = "$version" ]]; then
            echo "$cmd installed with acceptable version."
        else
            echo "The $cmd version is $current_version."
            echo "the $cmd need to be updated."
            echo "Updating..."

            apt-get upgrade ${cmd}

            echo "Successfully updated."
        fi
     else
        echo "The $cmd installed and up-to-date"
     fi

     echo ""
done < dependencies.txt

echo "All dependencies are installed and up-to-date."