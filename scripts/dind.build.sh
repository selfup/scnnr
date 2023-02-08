docker build -t scnnr_release . --build-arg CI=true --build-arg VERSION=($CI_PIPELINE_IID || (date +%s))

LATEST=$(docker run -td scnnr_release:latest)
ZIP_PATH="go/src/github.com/selfup/scnnr/scnnr_bins.zip"
SUM_PATH="go/src/github.com/selfup/scnnr/scnnr_bins.zip.sha256"

docker cp $LATEST:$ZIP_PATH scnnr_bins.zip
docker cp $LATEST:$SUM_PATH scnnr_bins.zip.sha256

ls scnnr_bins.zip || (echo 'no zip!' && exit 1)
ls scnnr_bins.zip.sha256 || (echo 'no sum!' && exit 1)

cat scnnr_bins.zip.sha256

echo "stopping $LATEST"

docker stop $LATEST

echo "release done"
