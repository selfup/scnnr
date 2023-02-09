set -e

go run cmd/pack/main.go

go run cmd/checksum/main.go

ls | grep scnnr

cat scnnr_bins.zip.sha256
