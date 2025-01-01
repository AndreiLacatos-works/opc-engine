#!/bin/bash

# Check if target bump (major, minor, or patch) is provided
if [ -z "$1" ]; then
  echo "Error: No version component specified. Use 'major', 'minor', or 'patch'."
  exit 1
fi

basedir=$(dirname "$0")

# Read the current version from versionfile
current_version=$(cat $basedir/versionfile)

# Split the version into major, minor, and patch components
IFS='.' read -r major minor patch <<< "$current_version"

# Increment the version
case $1 in
  major)
    major=$((major + 1))
    minor=0
    patch=0
    ;;
  minor)
    minor=$((minor + 1))
    patch=0
    ;;
  patch)
    patch=$((patch + 1))
    ;;
  *)
    echo "Error: Invalid argument. Use 'major', 'minor', or 'patch'."
    exit 1
    ;;
esac

# Construct the new version string
new_version="$major.$minor.$patch"

echo "Version bump: $current_version -> $new_version"

# Update the versionfile with the new version
echo "$new_version" > $basedir/versionfile
