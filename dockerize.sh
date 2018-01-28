#!/bin/bash

docker login -u "$DOCKERHUB_USERNAME" -p "$DOCKERHUB_PASSWORD"

cd $CI_PROJECT_DIR 

export VERSION=`cat VERSION`
if [ "$CI_COMMIT_REF_NAME" == "develop" ] then;
  VERSION = ${CI_COMMIT_SHA:0:8}
fi

docker build -t idstudios/clairctl:$VERSION .
docker tag idstudios/clairctl:$VERSION idstudios/clairctl:latest
docker push idstudios/clairctl:$VERSION
docker push idstudios/clairctl:latest
