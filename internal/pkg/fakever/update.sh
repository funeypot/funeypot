#!/usr/bin/env bash

set -e

curl -sSL 'https://raw.githubusercontent.com/openssh/openssh-portable/master/version.h' \
 | grep -E "^#define SSH_VERSION" \
 | awk -F "\"" '{print $2}' \
 | tr -d '\n' \
 > ssh.txt
