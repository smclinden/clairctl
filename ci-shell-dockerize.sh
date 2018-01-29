#!/bin/bash

IMAGE_NAME=idstudios/clairctl
docker login -u "$DOCKERHUB_USERNAME" -p "$DOCKERHUB_PASSWORD"

cd $CI_PROJECT_DIR 

export IMAGE_VERSION=`cat VERSION | ./version-inc.sh`
if [ "${CI_COMMIT_REF_NAME}" == "develop" ]; then
  IMAGE_VERSION=develop
fi

echo "***"
echo "*** Building Docker Image ${IMAGE_NAME}:${IMAGE_VERSION}"
echo "***"
docker build -t $IMAGE_NAME:$IMAGE_VERSION .

echo "***"
echo "*** Scanning Docker Image ${IMAGE_NAME}:${IMAGE_VERSION}"
echo "***"
./clairctl health
if [ $? -ne 0 ]; then
  echo ">>> Failed the ClairCtl health check!!!"

else
  ./clairctl analyze docker.io/$IMAGE_NAME:$IMAGE_VERSION --filters=Defcon1,Critical,High
  if [ $? -ne 0 ]; then
    echo ">>> Failed ClairCtl Vulnerability Criteia filters!!!"
    echo ">>> Running Detailed HTML Report"
    ./clairctl report docker.io/$IMAGE_NAME:$IMAGE_VERSION -f html
  else
    echo "***"
    echo "*** PASSED Clair Scanning Criteria!"
    echo "***"
  fi
fi 

docker push $IMAGE_NAME:$IMAGE_VERSION

if [ "${CI_COMMIT_REF_NAME}" == "master" ]; then
  echo "Marking image as latest"
  docker tag $IMAGE_NAME:$IMAGE_VERSION $IMAGE_NAME:latest
  docker push $IMAGE_NAME:latest
fi
