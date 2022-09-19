#!/bin/bash

set -xe
docker build -t mitaka8/playlist-bot:latest .
docker push mitaka8/playlist-bot:latest
