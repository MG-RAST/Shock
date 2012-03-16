#!/bin/sh

SHOCKURL=localhost:8000

PUT="curl -s -X PUT"
POST="curl -s -X POST"
GET="curl -s -X GET"
DELETE="curl -s -X DELETE"
#PP="python -mjson.tool"
PP="cat"

# create new node
# -> attributes
# -> data file
ATTR=""
DATA=""
echo "Create new node:"
$POST $ATTR $DATA "${SHOCKURL}/node" | $PP
echo

# list nodes
# -> limit &| skip
# -> query
L=""
S=""
Q=""
QUERY=$Q$L$S
echo "List nodes:"
$GET "${SHOCKURL}/node/${QUERY}" | $PP
echo

# get node
# download node file
ID="e841d01adf2ba939f034d55ea8e4b69d"
echo "Get node: ${ID}"
$GET "${SHOCKURL}/node/${ID}" | $PP
echo

