docker build -t scnnr_release . --build-arg CI=true --build-arg VERSION=$CI_PIPELINE_IID

$LATEST = $(docker run -td scnnr_release:latest)
$REL_PATH = "go/src/github.com/selfup/scnnr/scnnr_bins.zip"

docker cp "${LATEST}:${REL_PATH}" scnnr_bins.zip

(Get-ChildItem scnnr_bins.zip) -or (exit 1)

Write-Output "stopping $LATEST"

docker stop $LATEST

Write-Output "release done"
