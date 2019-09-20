#!/usr/bin/env bash

dockerfileFilename="Dockerfile"
goFileBedFilename="goFileBed"

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
read -p "please enter docker name(default:go_file_bed):" dockerName
if [ -z $dockerName ];then
    dockerName="go_file_bed"
fi

echo 'input any key go on,or control+c over'
read

if [ ! -f $dockerfileFilename ]; then
    wget -c -O $dockerfileFilename "https://raw.githubusercontent.com/cellargalaxy/goFileBed/master/Dockerfile"
else
    echo 'dockerfile exist'
fi
if [ ! -f $dockerfileFilename ]; then
    echo 'dockerfile not exist'
    exit 1
fi

if [ ! -f $goFileBedFilename ]; then
    wget -c -O $goFileBedFilename "https://github.com/cellargalaxy/goFileBed/releases/download/v0.2.0/goFileBed-linux"
else
    echo 'goFileBed exist'
fi
if [ ! -f $goFileBedFilename ]; then
    echo 'goFileBed not exist'
    exit 1
fi

echo 'chmod 755 '$goFileBedFilename
chmod 755 ./$goFileBedFilename

echo 'docker build'
docker build -t go_file_bed .
echo 'docker create volume'
docker volume create file_bed
echo 'docker run'
docker run -d --name $dockerName -v file_bed:/file_bed -p $listenPort:8880 -e TOKEN=$token -e LISTEN_ADDRESS=0.0.0.0:$listenPort go_file_bed

echo 'clear file'
rm -rf $dockerfileFilename
rm -rf $goFileBedFilename
echo 'clear file finish'

echo 'all finish'