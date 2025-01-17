#!/bin/bash

set -e

CVMFS_GATEWAY_SOURCES=$1
BUILD_PLATFORM=$2
VERSION=$3
RELEASE=$4

PROJECT_NAME=cvmfs-gateway

echo "Location: $CVMFS_GATEWAY_SOURCES"
echo "Platform: $BUILD_PLATFORM"
echo "Package version: $VERSION"
echo "Release: $RELEASE"

SCRIPT_LOCATION=$(cd "$(dirname "$0")"; pwd)
echo "Script location: $SCRIPT_LOCATION"

go_build_args=

if [ x"${BUILD_PLATFORM}" = xubuntu1604 ]; then
go_build_args="-ldflags=--extldflags=-Wl,--compress-debug-sections=none"
echo "Building on Ubuntu 16.04 with old ld: running go build with ${go_build_args}"
fi

echo "Building package"
cd ${CVMFS_GATEWAY_SOURCES}
export GOPATH=${GOPATH:=${CVMFS_GATEWAY_SOURCES}/../go}
export GOCACHE=${CVMFS_GATEWAY_SOURCES}/../gocache
go env
go build -mod=vendor $go_build_args

PACKAGE_NAME_SUFFIX="+$(lsb_release -si | tr [:upper:] [:lower:])$(lsb_release -sr)_amd64"
PACKAGE_NAME=cvmfs-gateway_$VERSION~$RELEASE$PACKAGE_NAME_SUFFIX.deb

mkdir -p ${CVMFS_GATEWAY_SOURCES}/DEBS

if [ -e /etc/profile.d/rvm.sh ]; then
    . /etc/profile.d/rvm.sh
fi

WORKSPACE=${CVMFS_GATEWAY_SOURCES}/pkg_ws_${BUILD_PLATFORM}
mkdir -p $WORKSPACE

mkdir -p $WORKSPACE/etc/systemd/system
mkdir -p $WORKSPACE/etc/cvmfs/gateway
mkdir -p $WORKSPACE/usr/bin
mkdir -p $WORKSPACE/usr/libexec/cvmfs-gateway/scripts
mkdir -p $WORKSPACE/var/lib/cvmfs-gateway

cp -v ${CVMFS_GATEWAY_SOURCES}/gateway $WORKSPACE/usr/bin/cvmfs_gateway

# Install the run_cvmfs_gateway.sh script for compatibility with cvmfs-gateway-1.0.0
cp -v ${CVMFS_GATEWAY_SOURCES}/pkg/run_cvmfs_gateway.sh ${WORKSPACE}/usr/libexec/cvmfs-gateway/scripts/

cp -v ${CVMFS_GATEWAY_SOURCES}/pkg/cvmfs-gateway.service \
    $WORKSPACE/etc/systemd/system/
cp -v ${CVMFS_GATEWAY_SOURCES}/pkg/cvmfs-gateway@.service \
    $WORKSPACE/etc/systemd/system/
cp -v ${CVMFS_GATEWAY_SOURCES}/config/repo.json $WORKSPACE/etc/cvmfs/gateway/
cp -v ${CVMFS_GATEWAY_SOURCES}/config/user.json $WORKSPACE/etc/cvmfs/gateway/

pushd $WORKSPACE
fpm -s dir -t deb \
    --verbose \
    --package ../DEBS/$PACKAGE_NAME \
    --version $VERSION \
    --name cvmfs-gateway \
    --maintainer "Radu Popescu <radu.popescu@cern.ch>" \
    --description "CernVM-FS Repository Gateway" \
    --url "http://cernvm.cern.ch" \
    --license "BSD-3-Clause" \
    --depends "cvmfs-server > 2.5.2" \
    --replaces "cvmfs-notify" \
    --directories usr/libexec/cvmfs-gateway \
    --config-files etc/cvmfs/gateway/repo.json \
    --config-files etc/cvmfs/gateway/user.json \
    --config-files etc/systemd/system/cvmfs-gateway.service \
    --config-files etc/systemd/system/cvmfs-gateway@.service \
    --exclude etc/systemd/system \
    --no-deb-systemd-restart-after-upgrade \
    --after-install ${CVMFS_GATEWAY_SOURCES}/pkg/setup_deb.sh \
    --chdir $WORKSPACE \
    ./
popd

mkdir -p ${CVMFS_GATEWAY_SOURCES}/pkgmap
PKGMAP_FILE=${CVMFS_GATEWAY_SOURCES}/pkgmap/pkgmap.${BUILD_PLATFORM}_x86_64
echo "[${BUILD_PLATFORM}_x86_64]" >> $PKGMAP_FILE
echo "gateway=$PACKAGE_NAME" >> $PKGMAP_FILE

rm -rf $WORKSPACE
