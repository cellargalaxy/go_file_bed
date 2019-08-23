#!/usr/bin/env bash

dockerfileFilename="Dockerfile"
goFileBedConfigFilename="config.yml"
goFileBedFilename="goFileBed-linux"

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
read -p "please enter synUrl(default:http://127.0.0.1:8880):" synUrl
if [ -z $synUrl ];then
    synUrl="http://127.0.0.1:8880"
fi
goFileBedConfig='token: '$token'
synUrl: '$synUrl'
listenAddress: 0.0.0.0:8880
fileBedPath: file_bed'

echo 'input any key go on,or control+c over'
read
cat>$goFileBedConfigFilename<<EOF
$goFileBedConfig
EOF

wget -c -O $dockerfileFilename "https://raw.githubusercontent.com/cellargalaxy/goFileBed/master/Dockerfile"

wget -c -O $goFileBedFilename "https://github.com/cellargalaxy/goFileBed/releases/download/v0.1.3/goFileBed-linux"

if [ ! -f $dockerfileFilename ]; then
    echo 'Dockerfile not exist'
    exit 1
fi
if [ ! -f $goFileBedConfigFilename ]; then
    echo 'config not exist'
    exit 1
fi
if [ ! -f $goFileBedFilename ]; then
    echo 'goFileBed not exist'
    exit 1
fi

echo 'chmod 755 goFileBed'
chmod 755 ./$goFileBedFilename

echo 'docker build'
docker build -t go_file_bed .
echo 'docker create volume'
docker volume create file_bed
echo 'docker run'
docker run -d --name go_file_bed -v file_bed:/file_bed -p $listenPort:8880 go_file_bed

echo 'clear file'
rm -rf $dockerfileFilename
rm -rf $goFileBedConfigFilename
rm -rf $goFileBedFilename
echo 'clear file finish'

echo 'all finish'