#!/bin/bash

GITHUB_USER=ids
GITHUB_REPO=clairctl
RELEASE_DESC="A CLI based utility for interacting with the Clair API"
RELEASE_TITLE=" - stable release"
mkdir -p $GOPATH/src/$(dirname $REPO_NAME)
ln -svf $CI_PROJECT_DIR $GOPATH/src/$REPO_NAME
cd $GOPATH/src/$REPO_NAME
curl https://glide.sh/get | sh
glide install -v
go get -u github.com/jteeuwen/go-bindata/...
go generate ./clair
go get github.com/mitchellh/gox
gox -os="linux" -os="darwin" -arch="amd64" -output="client-bins/{{.Dir}}-{{.OS}}-{{.Arch}}" -ldflags "-X github.com/ids/clairctl/cmd.version=$(cat VERSION)"
go get github.com/aktau/github-release

cd $CI_PROJECT_DIR

git config --global user.email $GITHUB_EMAIL
git config --global user.name $GITHUB_USERNAME

echo "configured remotes:"
git remote -v
git remote remove github
git remote add github "https://$GITHUB_USERNAME:$GITHUB_TOKEN@github.com/ids/clairctl"

echo "re-configured remote w/ token:"
git remote -v

export VERSION=develop
export PRE_FLAG=
echo "CI_COMMIT_REF_NAME: ${CI_COMMIT_REF_NAME}"

if [ "${CI_COMMIT_REF_NAME}" == "develop" ]; then
  PRE_FLAG=--pre-release
  RELEASE_TITLE=" - unstable"
else
  # it's master, update the version
  VERSION=`cat VERSION`
  VERSION=`echo $VERSION | ./version-inc.sh` 
fi

THEME=$(git log -1 --pretty=%B)

echo "THEME: ${THEME}"
echo "VERSION: ${VERSION}"
echo "creating tag ${VERSION}"
git tag -a $VERSION -m "${VERSION}" -f 
git push --force github refs/tags/${VERSION}:refs/tags/${VERSION}

github-release release \
  --user $GITHUB_USER \
  --repo $GITHUB_REPO \
  --tag $VERSION \
  --name "${VERSION} ${RELEASE_TITLE}" \
  --description "${RELEASE_DESC}" \
  $PRE_FLAG  

if [ $? -ne 0 ]; then
  github-release delete \
    --user $GITHUB_USER \
    --repo $GITHUB_REPO \
    --tag $VERSION \

  github-release release \
    --user $GITHUB_USER \
    --repo $GITHUB_REPO \
    --tag $VERSION \
    --name "${VERSION} ${RELEASE_TITLE}" \
    --description "${RELEASE_DESC}" \
    $PRE_FLAG  
fi

github-release upload \
    --user $GITHUB_USER \
    --repo $GITHUB_REPO \
    --tag $VERSION \
    --replace \
    --name "clairctl-darwin-amd64" \
    --file $CI_PROJECT_DIR/client-bins/clairctl-darwin-amd64

github-release upload \
    --user $GITHUB_USER \
    --repo $GITHUB_REPO \
    --tag $VERSION \
    --replace \
    --name "clairctl-linux-amd64" \
    --file $CI_PROJECT_DIR/client-bins/clairctl-linux-amd64


if [ "$CI_COMMIT_REF_NAME" == "master" ]; then
  rm -rf client-bins
  git status
  git checkout master
  git pull github master
  git status
  git branch
  echo $VERSION > VERSION
  git add VERSION
  git commit -a -m "automated update to version: ${VERSION}"
  git push github master
fi

echo "Release binairies have been deployed and updated"
