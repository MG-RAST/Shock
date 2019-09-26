#!/bin/sh -e

# /testdata/attr.json
# /testdata/10kb.fna
# /testdata10kb.fna.gz
# /testdata/40kb.fna
# /testdata/sample1.fq
# /testdata/nr_subset1.fa
# /testdata/nr_subset2.fa
# /testdata/nr_subset.tar.gz

HOST=http://localhost
PORT=7445
TOKEN=""

while getopts h:p:t: option; do
    case "${option}"
        in
            h) HOST=${OPTARG};;
            p) PORT=${OPTARG};;
            t) TOKEN=${OPTARG};;
    esac
done

export SHOCK_URL=$HOST:$PORT
if [ ! -z "$TOKEN" ]; then
    export TOKEN=$TOKEN
fi

NODEID=""
SC=/go/bin/shock-client

run_test () {
    CMD=$1
    echo "Running: $CMD"
    RESULT=`$CMD`
    EXIT=$?
    if [ $EXIT -ne 0 ]; then
        echo "Failed: $CMD"
        exit 1
    fi
    # create and update produces node id
    if [ -z "$2" ]; then
        NODEID=""
    else
        NODEID=`echo $RESULT | tail -1 | cut -f3 -d' '`
    fi
}

edit_attr () {
    NAME=$1
    FORM=$2
    FILE=$3
    sed -e "s/replace_name/$NAME/" -e "s/replace_format/$FORM/" /testdata/attr.json > $FILE
}

# run tests

run_test "$SC info"

run_test "$SC create --filepath /testdata/10kb.fna" "ID"
ID1=$NODEID
run_test "$SC get $ID1"
run_test "$SC index $ID1 line"
run_test "$SC download --md5 $ID1"

edit_attr "chunk" "fasta" "/testdata/a2"
run_test "$SC create --chunk 5K --filepath /testdata/40kb.fna --attributes /testdata/a2" "ID"
ID2=$NODEID
run_test "$SC index $ID2 record"
sleep 5 # wait in index action
run_test "$SC download --index record --parts 10-20 $ID2"

edit_attr "parts" "fasta" "/testdata/a3"
run_test "$SC create --part 2 --attributes /testdata/a3" "ID"
ID3=$NODEID
run_test "$SC update --part 1 --filepath /testdata/nr_subset1.fa $ID3"
run_test "$SC update --part 2 --filepath /testdata/nr_subset2.fa $ID3"
run_test "$SC update --filename foo.bar $ID3"
run_test "$SC get $ID3"

edit_attr "gzip" "fasta" "/testdata/a4"
run_test "$SC create --compression gzip --filepath /testdata/10kb.fna.gz --attributes /testdata/a4" "ID"
ID4=$NODEID

edit_attr "unpack" "fasta" "/testdata/a5"
run_test "$SC create --filepath /testdata/nr_subset.tar.gz" "ID"
ID5=$NODEID
run_test "$SC unpack --archive tar.gz $ID5 --attributes /testdata/a5"

edit_attr "copy" "fasta" "/testdata/a6"
run_test "$SC create --copy $ID1 --attributes /testdata/a6" "ID"
ID6=$NODEID
run_test "$SC delete $ID1"

edit_attr "update" "fastq" "/testdata/a7"
run_test "$SC create" "ID"
ID7=$NODEID
run_test "$SC update --filepath /testdata/sample1.fq --attributes /testdata/a7 $ID7"

# distinct query required mongodb index on field
# run_test "$SC query --distinct name"
run_test "$SC query --attribute format:fasta --attribute name:unpack"
run_test "$SC query --other file.name:nr_subset1.fa"
run_test "$SC query --limit 5"

