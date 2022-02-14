#!/bin/bash

set -ev

sudo apt update && sudo apt install ca-certificates libgnutls30 -y
sudo apt-get -qq -y install conntrack socat
