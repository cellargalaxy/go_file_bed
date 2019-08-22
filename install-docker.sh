#!/usr/bin/env bash

dockerfileFilename="Dockerfile"
goFileBedConfigFilename="config.yml"
goFileBedFilename="goFileBed-linux"

while :
do
    read -p "please enter token(required):" token
    if [ ! -z $token ];then
        break
    fi
done
read -p "please enter listen port(default:8880):" listenPort
if [ -z $listenPort ];then
    listenPort="8880"
fi
goFileBedConfig='token: '$token'
synUrl: http://127.0.0.1:8880
listenAddress: 0.0.0.0:8880
fileBedPath: file_bed'

echo $goFileBedConfig
echo 'config,input any key go on,or control+c over'
read
echo $goFileBedConfig > $goFileBedConfigFilename

echo `wget -q -O - https://raw.githubusercontent.com/cellargalaxy/goFileBed/master/Dockerfile` > $dockerfileFilename

wget -O $goFileBedFilename "https://github.com/cellargalaxy/goFileBed/releases/download/v0.1.1/goFileBed-linux"
chmod 755 ./$goFileBedFilename

docker build -t go_file_bed .
docker run -d --name go_file_bed -p $listeningPort:8880 go_file_bed

rm -rf $dockerfileFilename
rm -rf $goFileBedConfigFilename
rm -rf $goFileBedFilename