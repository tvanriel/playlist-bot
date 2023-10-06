#!/bin/bash

set -ex


source="$1"
dbPort="$2"
dbName="$3"
dbUser="$4"
dbPass="$5"

function get_s3_files() {
        mcli ls $1 --json | jq '.key' -r | tr -d '/' | sort > /tmp/s3_track_entries.txt
}

function get_artist() {
        fn=$(echo "$1/$2" | sed "s/$/\/metadata.json/")
        mcli cat $fn | jq .Artist -r
}

function get_trackname() {
        fn=$(echo "$1/$2" | sed "s/$/\/metadata.json/")
        mcli cat $fn | jq .Name -r
}

function update_trackname() {
        port=$1
        name=$2
        user=$3
        pass=$4
        uuid=$5
        artist=$(echo $6 | sed "s/'//g" | sed "s/\"//g")
        track=$(echo $7 | sed "s/'//g" | sed "s/\"//g")

        /usr/bin/mycli -u "$user" -P "$port" -p "$pass" "$name" -e "UPDATE track_models set artist_name = '$artist', track_name = '$track' WHERE uuid = '$uuid' LIMIT 1;"
}


get_s3_files $1


while read t; do
    echo $t
    artist=$(get_artist $1 $t)
    trackname=$(get_trackname $1 $t)
    update_trackname $dbPort $dbName $dbUser $dbPass "$t" "$artist" "$trackname"
done </tmp/s3_track_entries.txt






