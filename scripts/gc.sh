#!/bin/bash

function get_s3_files() {
        mcli ls flagship/playlists/972617390816915516/ --json | jq '.key' -r | tr -d '/' | sort > /tmp/s3_track_entries.txt
}
function get_mysql_files() {
        mycli --csv -uroot -proot -P 33060 playlists -e 'select uuid from track_models where deleted_at IS NULL' | sort | sed 's/"//g' |sed '/^uuid$/d' > /tmp/mysql_track_entries.txt
}

function get_files_in_s3_not_in_mysql() {
        diff --new-line-format="" --unchanged-line-format="" /tmp/s3_track_entries.txt /tmp/mysql_track_entries.txt | sed '/^$/d' > s3_mysql.txt
}
function get_files_in_mysql_not_in_s3() {
        diff --new-line-format="" --unchanged-line-format="" /tmp/mysql_track_entries.txt /tmp/s3_track_entries.txt | sed '/^$/d' > mysql_s3.txt
}


function remove_s3_files() {
        cat s3_mysql.txt | xargs -I '{}' mcli rm --recursive --force flagship/playlists/972617390816915516/{}/
}
function remove_mysql_files() {
        cat mysql_s3.txt |  xargs -I '{}' mycli -P 33060 -proot -uroot playlists -e 'UPDATE track_models set deleted_at=NOW() where uuid = "{}"'
}

get_mysql_files
get_s3_files

get_files_in_mysql_not_in_s3
get_files_in_s3_not_in_mysql

#remove_s3_files
#remove_mysql_files
