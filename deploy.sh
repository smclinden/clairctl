#!/bin/bash

GITHUB_USER=ids
GITHUB_REPO=clairctl
RELEASE_DESC="A CLI based utility for interacting with the Clair API"

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
echo $PWD
echo $CI_COMMIT_TAG
VERSION=`cat VERSION`_${CI_COMMIT_SHA:0:8}
echo "VERSION: ${VERSION}"

git config --global user.email $GITHUB_EMAIL
git config --global user.name $GITHUB_USERNAME

echo "configured remotes:"
git remote -v
git remote remove github
git remote add github "https://$GITHUB_USERNAME:$GITHUB_TOKEN@github.com/ids/clairctl"

echo "re-configured remote w/ token:"
git remote -v

echo "creating tag ${VERSION}"
git tag -a $VERSION -m "${VERSION}" -f 
git push github --tags

github-release release \
  --user $GITHUB_USER \
  --repo $GITHUB_REPO \
  --tag $VERSION \
  --name "${VERSION} Release" \
  --description "${RELEASE_DESC}" \
  --pre-release  

github-release upload \
    --user $GITHUB_USER \
    --repo $GITHUB_REPO \
    --tag $VERSION \
    --name "clairctl-darwin-amd64" \
    --file $CI_PROJECT_DIR/client-bins/clairctl-darwin-amd64

github-release upload \
    --user $GITHUB_USER \
    --repo $GITHUB_REPO \
    --tag $VERSION \
    --name "clairctl-linux-amd64" \
    --file $CI_PROJECT_DIR/client-bins/clairctl-linux-amd64

echo "Release binairies have been deployed and updated"
