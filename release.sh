#! /bin/sh

set -ex

VERSION=0.0.1

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
cd $SCRIPT_DIR

# build, tag the image
docker build -t totp:${VERSION} -f Dockerfile .
docker tag totp:${VERSION} uristmcdwarf/sidecar:${VERSION}

# set latest tag
docker tag totp:${VERSION} uristmcdwarf/sidecar:latest

# push the image
docker push uristmcdwarf/sidecar:${VERSION}
docker push uristmcdwarf/sidecar:latest

cd -
