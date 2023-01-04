#!/bin/bash

if [ ! -f "backend.conf" ]; then
    echo "Error: backend.conf does not exist. Please refer to README.md for more information"
    exit 1
fi

if [ ! -f "config.yaml" ]; then
    echo "Error: config.yaml does not exist. Please refer to README.md for more information"
    exit 1
fi

cd platform
terraform init -backend-config=../backend.conf

ENV="production"
terraform workspace select $ENV
terraform plan -out $ENV.tfplan
terraform apply "$ENV.tfplan"
