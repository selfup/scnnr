docker build -t scnnr_release . \
  --build-arg CI=true \
  --build-arg VERSION=$CI_PIPELINE_IID

LATEST=$(docker run -td scnnr_release:latest)
REL_PATH="go/src/github.com/selfup/scnnr/scnnr_bins.zip"

docker cp $LATEST:$REL_PATH scnnr_bins.zip

ls scnnr_bins.zip || exit 1

echo "stopping $LATEST"

docker stop $LATEST

echo "stopped"

chmod +x scnnr_bins.zip

echo "release done"
