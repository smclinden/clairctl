#!/bin/bash

IMG_NAME=idstudios/clairctl
docker login -u "$DOCKERHUB_USERNAME" -p "$DOCKERHUB_PASSWORD"

cd $CI_PROJECT_DIR 

export VERSION=`echo $VERSION | ./version-inc.sh`
if [ "$CI_COMMIT_REF_NAME" == "develop" ]; then
  VERSION = "develop"
fi

echo "Building Docker Image $IMAGE_NAME"
docker build -t $IMG_NAME:$VERSION .
docker tag $IMAGE_NAME:$VERSION $IMAGE_NAME:latest
docker push $IMAGE_NAME:$VERSION
docker push $IMAGE_NAME:latest
