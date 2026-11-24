#!/bin/sh
set -e 

ROOT_DIR="$(pwd)"

if [ -d "worker/build" ]; then
  cd worker/build
else
  cd worker
  cmake -B build/ -G "Unix Makefiles" -DCMAKE_EXPORT_COMPILE_COMMANDS=ON -DCMAKE_BUILD_TYPE=Debug . 
  cd build
fi

make

mkdir -p "$ROOT_DIR/bin"

cp TaskforgeWorker "$ROOT_DIR/bin/TaskForge-Worker"

