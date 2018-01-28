#!/bin/bash

GITHUB_USER=ids
GITHUB_REPO=clairctl

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
VERSION=`cat $CI_PROJECT_DIR/VERSION`

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
