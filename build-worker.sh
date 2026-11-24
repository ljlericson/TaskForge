#!/bin/sh
set -e 

ROOT_DIR="$(pwd)"

if [ -d "worker/build" ]; then
  cd worker/build
else
  cd worker
  sh generate-makefile.sh
  cd build
fi

make

mkdir -p "$ROOT_DIR/bin"

cp TaskforgeWorker "$ROOT_DIR/bin/TaskForge-Worker"
