#!/bin/sh
set -eu

awslocal sqs create-queue --queue-name product-events >/dev/null
