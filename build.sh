#!/usr/bin/env bash

TAG=${1:-latest}
IMG="agill17/s3-operator"

echo "########################################################################################"
echo "                           BUILDING -> ${IMG}:${TAG}"
echo "########################################################################################"


set -e

# generate openapi and crds
operator-sdk0.15.2 generate openapi

# copy generated crd to chart
cp deploy/crds/agill.apps_s3s_crd.yaml chart/s3-operator/templates/

# set new version in charts values
sed -E -i "" "s~${IMG}:.*~${IMG}:${TAG}~g" chart/s3-operator/values.yaml


## go fmt
go fmt ./pkg/...
go fmt ./cmd/...

# build and push
operator-sdk0.15.2 build $IMG:$TAG && docker push $IMG:$TAG
if [[ $TAG != "latest" ]]
then
    echo "########################################################################################"
    echo "                             TAGGING LATEST IMAGE                                       "
    echo "########################################################################################"
    docker tag $IMG:$TAG $IMG:latest
    docker push $IMG:latest
fi

echo "########################################################################################"
echo "                                  NEW CHANGES                                           "
echo "########################################################################################"
# git diff
git --no-pager diff