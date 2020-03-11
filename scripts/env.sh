#!/bin/bash -e

MONGO_ADDR=localhost
echo Setting mongodb address to $MONGO_ADDR
export MONGODB_URI=mongodb://$MONGO_ADDR:27017
