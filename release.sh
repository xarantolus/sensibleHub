mkdir -p releases
rm releases/*

GIT_COMMIT=$(git rev-parse --short HEAD)

GOOS=windows ./pack.sh "releases/sensibleHub-windows-$GIT_COMMIT.zip"

GOOS=linux ./pack.sh "releases/sensibleHub-linux-$GIT_COMMIT.zip"

GOOS=linux GOARCH=arm GOARM=7 ./pack.sh "releases/sensibleHub-raspberrypi-$GIT_COMMIT.zip"