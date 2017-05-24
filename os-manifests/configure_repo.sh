#!/bin/sh

if [ $# -ne 2 ]; then
  echo "USAGE: configure_img.sh [OLD_IMG_REPO] [NEW_IMG_REPO]"
  echo "NOTE: in case old_repo is wrong, the replacement does nothing"
  exit 1
fi
echo "configure_img.sh $1 $2"

sed -ri 's/image: '${1}'/image: '${2}'/g' engine/engine-deployment.yaml vdsc/vdsc-deployment.yaml
echo ""
echo ""
echo "Grepping images after modification:"
echo ""
grep -hr "image:" engine/* vdsc/*
