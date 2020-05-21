#!/bin/sh
# test round trip local file to SHOCK, S3 and back, end with diff

# 
cp $1 /tmp/lolo.temp

# upload a file to shock
ID=$(curl -s -X POST -F "upload=@/tmp/lolo.temp" http://localhost:7445/node |  jq .data.id | sed 's/"//g' ) 

echo "ID=${ID}"

curl -s -X GET http://localhost:7445/node/${ID}?download > /tmp/lala.temp 

diff -c /tmp/lolo.temp /tmp/lala.temp 
if [ $? -ne 0 ] ; then
    echo "files differ in first test"
    exit 1
else
    echo "Passed first test"
fi

# remove re-downloaded copy
rm -f tmp/lala.temp

# #################################################################
# copy file to local s3
mc mb s3/bucket1 > /dev/null 2>&1
mc --quiet cp /tmp/lolo.temp s3/bucket1/${ID}.data  > /dev/null 2>&1

echo "adding location info for local minio based S3"
RES1=$(curl -s -X POST -H 'Authorization: basic YWRtaW46c2VjcmV0' -H "Content-Type: application/json" "http://localhost:7445/node/${ID}/locations/" -d '{"id":"s3"}' | jq .status )

if [ "$RES1"x != "200x" ] ; then
    echo "return != 200 [$RES1] "
    exit 1
fi

echo "find the location that is stored on the server"

RES2=$(curl -s -X GET http://localhost:7445/node/${ID} |  jq .data.locations[].id )
echo "Checking location stored. Found ${RES2}"

# #################################################################

#  substr( $0,0,2) "/"  substr( $0,3,2) "/" substr( $0,5,2) "/" $0 "/" $0 ".data
path=$(echo $ID | awk ' { print "/usr/local/shock/data/" substr( $0,0,2) "/"  substr( $0,3,2) "/" substr( $0,5,2) "/" $0 "/" $0 ".data" } ' )
echo "Now go ahead and stop the server, delete the file in "
echo " "
echo "  ${path}"
echo " "
echo " press ANY key to proceed"

read -n 1 -s

RES2=$(curl -s -X GET http://localhost:7445/node/${ID} |  jq .data.locations[].id )
echo "Checking location stored. Found ${RES2}"


echo "attempting to download"
curl -s -X GET http://localhost:7445/node/${ID}?download > /tmp/lala.temp

echo "comparing files downloaded "
diff -c /tmp/lolo.temp /tmp/lala.temp 
if [ $? -ne 0 ] ; then
  echo "files differ in second attempt"
  exit 1
fi

echo "All tests successful"

# remove all
rm -f /tmp/lolo.temp /tmp/lala.temp
