#!/bin/bash

# define the path to the version file
VERSION_FILE="src/versionfile"

# check if the version file exists
if [[ ! -f "$VERSION_FILE" ]]; then
    echo "Error: Version file not found at $VERSION_FILE."
    exit 1
fi

# read the contents of the version file
CURRENT_VERSION=$(<"$VERSION_FILE")

# prompt the user
echo "Current version is $CURRENT_VERSION. Do you want to upgrade? Y/n"

# read user input
read -r response

# handle user response
if [[ "$response" =~ ^[Yy]$ || -z "$response" ]]; then
    # ensure version bump utility exists & executable
    VERSION_BUMP_UTIL="src/versionbump.sh"
    if [[ ! -f "$VERSION_BUMP_UTIL" ]]; then
        echo "Error: Version bump utility found at $VERSION_BUMP_UTIL"
        exit 1
    fi
    chmod +x $VERSION_BUMP_UTIL

    # prompt to select the version component to bump
    echo "Select version component to bump:"
    echo -e "\t1 - major"
    echo -e "\t2 - minor"
    echo -e "\t3 - patch"

    # read user input
    read -r component

    # handle user input and validate selection
    case "$component" in
        1)
            echo "Upgrading major"
            bash "$VERSION_BUMP_UTIL" "major"
            ;;
        2)
            echo "Upgrading minor"
            bash "$VERSION_BUMP_UTIL" "minor"
            ;;
        3)
            echo "Upgrading patch"
            bash "$VERSION_BUMP_UTIL" "patch"
            ;;
        *)
            echo "Invalid selection. Please choose 1, 2, or 3."
            exit 1
            ;;
    esac

    CURRENT_VERSION=$(<"$VERSION_FILE")

    echo "Proceeding with new version: $CURRENT_VERSION"
else
    echo "Proceeding with version $CURRENT_VERSION"
fi

echo "Building docker image"
docker build --no-cache -t opc-engine-simulator:"$CURRENT_VERSION" .

# create an archive of the image
mkdir -p dist
docker save -o dist/opc-engine-simulator-v$CURRENT_VERSION.tar opc-engine-simulator:$CURRENT_VERSION

# update docker-compose
sed -i "s/image: opc-engine-simulator:.*$/image: opc-engine-simulator:${CURRENT_VERSION}/g" docker-compose.yaml
