#!/bin/bash
CONFIG_PATH="$1"

ACCESS_KEY="$(jq -r .credential.accessKey $CONFIG_PATH)"
SECRET_KEY="$(jq -r .credential.secretKey $CONFIG_PATH)"
ARN="arn:minio:sqs:us-east-1:1:webhook"

URL="http://localhost:9000"
ALIAS="local-minio"
BUCKET="layers"

function check_configuration() {
    local s=$(mc config host list | grep $ALIAS | wc -l)
    test $s -eq 1
}

function configure() {
    mc config host add $ALIAS $URL $ACCESS_KEY $SECRET_KEY
}


function check_bucket() {
    mc ls $ALIAS/$BUCKET
}

function create_bucket() {
    mc mb $ALIAS/$BUCKET
}



function event_handler_ok() {
    local s=$(mc events list $ALIAS/$BUCKET $ARN | wc -l)
    test $s -eq 1
}

function add_event_handler() {
    mc events add $ALIAS/$BUCKET "$ARN" --events put
}


check_configuration || configure
check_bucket        || create_bucket
event_handler_ok    || add_event_handler
