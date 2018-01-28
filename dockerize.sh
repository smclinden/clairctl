#!/bin/bash

IMAGE_NAME=idstudios/clairctl
docker login -u "$DOCKERHUB_USERNAME" -p "$DOCKERHUB_PASSWORD"

cd $CI_PROJECT_DIR 

export VERSION=`cat VERSION | ./version-inc.sh`
if [ "$CI_COMMIT_REF_NAME" == "develop" ]; then
  VERSION = "develop"
  echo "Building Docker Image ${IMAGE_NAME}:${VERSION}"
  docker build -t $IMAGE_NAME:$VERSION .
  docker push $IMAGE_NAME:$VERSION
else
  echo "Building Docker Image ${IMAGE_NAME}:${VERSION} as latest release"
  docker build -t $IMAGE_NAME:$VERSION .
  docker tag $IMAGE_NAME:$VERSION $IMAGE_NAME:latest
  docker push $IMAGE_NAME:$VERSION
  docker push $IMAGE_NAME:latest
fi
