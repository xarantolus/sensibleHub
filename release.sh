mkdir -p releases
rm releases/*

chmod +x pack.sh

GIT_COMMIT=$(git rev-parse --short HEAD)

GOOS=windows ./pack.sh "releases/sensibleHub-windows-$GIT_COMMIT.zip"

GOOS=linux ./pack.sh "releases/sensibleHub-linux-$GIT_COMMIT.zip"

GOOS=linux GOARCH=arm GOARM=5 ./pack.sh "releases/sensibleHub-raspberrypi-$GIT_COMMIT.zip"

# Generate Changelog
git log $(git describe --tags --abbrev=0)..HEAD --oneline | cut -d" " -f2- - | awk '$0="* "$0'  > release.md
