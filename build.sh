#!/bin/bash

IMAGE_NAME="javafuns/servicecache-operator:v0.0.1"

operator-sdk build ${IMAGE_NAME}

sed -i "s|REPLACE_IMAGE|${IMAGE_NAME}|g" deploy/operator.yaml

docker push ${IMAGE_NAME}

sed -i "s|${IMAGE_NAME}|REPLACE_IMAGE|g" deploy/operator.yaml
