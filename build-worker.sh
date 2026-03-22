if [ -d "worker/build" ]; then
  cd worker/build
  make
  cp TaskforgeWorker ../../bin/TaskForge-Worker
else
  cd worker
  sh generate-makefile.sh
  cd build
  make
  cp TaskforgeWorker ../../bin/TaskForge-Worker
fi
