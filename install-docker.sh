#!/usr/bin/env bash

while :
do
    read -s -p "please enter token(required):" token
    if [ ! -z $token ];then
        break
    fi
done
read -p "please enter listen port(default:8880):" listenPort
if [ -z $listenPort ];then
    listenPort="8880"
fi
read -p "please enter last file info count(default:10):" lastFileInfoCount
if [ -z $lastFileInfoCount ];then
    lastFileInfoCount="10"
fi

echo 'input any key go on,or control+c over'
read

echo 'docker build'
docker build -t go_file_bed .
echo 'docker create volume'
docker volume create go_file_bed
echo 'docker run'
docker run -d --restart=always --name go_file_bed -p $listenPort:8880 -e TOKEN=$token -e LAST_FILE_INFO_COUNT=$lastFileInfoCount -v file_bed:/file_bed go_file_bed

echo 'all finish'