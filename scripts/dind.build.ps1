docker build -t scnnr_release . --build-arg CI=true --build-arg VERSION=$CI_PIPELINE_IID

$LATEST = $(docker run -td scnnr_release:latest)
$ZIP_PATH = "go/src/github.com/selfup/scnnr/scnnr_bins.zip"
$SUM_PATH = "go/src/github.com/selfup/scnnr/scnnr_bins.zip.sha256"

docker cp "${LATEST}:${ZIP_PATH}" scnnr_bins.zip
docker cp "${LATEST}:${SUM_PATH}" scnnr_bins.zip.sha256

(Get-ChildItem scnnr_bins.zip) -or (Write-Output "no zip!" -and exit 1)
(Get-ChildItem scnnr_bins.zip.sha256) -or (Write-Output "no sum!" -and exit 1)

Write-Output "stopping $LATEST"

docker stop $LATEST

Write-Output "release done"
