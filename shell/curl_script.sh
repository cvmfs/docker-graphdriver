#!/bin/bash
REPO=library/ubuntu
TAG=latest
REGISTRY=https://registry-1.docker.io/v2

#URI="$REGISTRY/$REPO/manifests/$TAG"
URI="$REGISTRY/$REPO/tags/list"
echo URI=$URI
MANIFEST="`curl -skL -o /dev/null -D- $URI`"
CHALLENGE="`grep "Www-Authenticate" <<<"$MANIFEST"`"
echo "$CHALLENGE"
if [[ CHALLENGE ]]; then
    IFS=\" read _ REALM _ SERVICE _ SCOPE _ <<<"$CHALLENGE"
    echo "IFS is: " "$IFS"
    echo REALM is $REALM
    echo SERVICE is $SERVICE
    echo SCOPE is $SCOPE
    TOKEN="`curl -skL "$REALM?service=$SERVICE&scope=$SCOPE"`"
    IFS=\" read _ _ _ TOKEN _ <<<"$TOKEN"
    echo TOKEN is $TOKEN
    MANIFEST="`curl -isk -X GET -H "Authorization: Bearer $TOKEN" $URI`"
    echo "RESPONSE is $MANIFEST"
fi
