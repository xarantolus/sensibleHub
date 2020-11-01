mkdir -p releases
rm releases/*

chmod +x pack.sh

GIT_COMMIT=$(git rev-parse --short HEAD)

GOOS=windows ./pack.sh "releases/sensibleHub-windows-$GIT_COMMIT.zip"

GOOS=linux ./pack.sh "releases/sensibleHub-linux-$GIT_COMMIT.zip"

GOOS=linux GOARCH=arm GOARM=5 ./pack.sh "releases/sensibleHub-raspberrypi-$GIT_COMMIT.zip"
