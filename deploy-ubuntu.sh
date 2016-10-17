#!/bin/bash

sudo apt-get update
sudo apt install -y golang-go

git clone https://github.com/bmeg/task-execution-server.git
git checkout scaling

make depends
make
