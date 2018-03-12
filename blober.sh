# This is the universal installer for the blobfile utility.

BLOB_LATEST_VERSION=""

if [[ $# -ne 1 && -z "$BLOB_LATEST_VERSION" ]]; then
    echo "Must provide a version to download"
    exit -1
elif [ $# -eq 1 ]; then
    BLOB_LATEST_VERSION=$1
fi

DOWNLOAD_FILE=$(mktemp)
UNZIP_DIR=$(mktemp -d)

curl "https://aleem.haji.ca/blob/clientlib/$BLOB_LATEST_VERSION.zip" -o "$DOWNLOAD_FILE"
unzip -o "$DOWNLOAD_FILE" -d "$UNZIP_DIR"
cd "$UNZIP_DIR"
make install
