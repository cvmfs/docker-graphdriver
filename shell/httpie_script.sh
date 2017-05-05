#!/bin/bash
REPO="library/ubuntu"
REGISTRY="https://registry-1.docker.io/v2"
BLOB="sha256:5f70bf18a086007016e948b04aed3b82103a36bea41755b6cddfaf10ace3c6ef"
URI="$REGISTRY/$REPO/manifests/latest"
#URI="$REGISTRY/$REPO/blobs/$BLOB"
SCHEMA2_HDR="application/vnd.docker.distribution.manifest.v2+json"
echo "URI is: $URI"

RESPONSE="`http --headers \"$URI\"`"
echo "Response is: $RESPONSE"

CHALLENGE="`grep "Www-Auth" <<<"$RESPONSE"`"
echo "Challenge is: $CHALLENGE"

IFS=\" read _ REALM _ SERVICE _ SCOPE _ <<<"$CHALLENGE"
echo "Realm is: $REALM"
echo "Service is: $SERVICE"
echo "Scope is: $SCOPE"

TOKEN="`http --body "$REALM?service=$SERVICE&scope=$SCOPE" | jq ".token"`"
TOKEN="$(sed s/\"//g <<<$TOKEN)"
echo "Token is: $TOKEN"

http -v --follow "$URI" Accept:"$SCHEMA2_HDR" Authorization:"Bearer $TOKEN" 

echo "-----------"
