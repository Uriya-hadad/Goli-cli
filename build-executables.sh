#!/usr/bin/env bash
latest_version=$(grep -m 1 -oE "\[[0-9]+\.[0-9]+\.[0-9]+\]" CHANGELOG.md | sed 's/\[\(.*\)\]/\1/')

platforms=(
#"darwin/amd64"
"darwin/arm64"
#"linux/amd64"
#"linux/arm"
#"linux/arm64"
"windows/amd64"
)

for platform in "${platforms[@]}"
do
  platform_split=(${platform//\// })
  GOOS=${platform_split[0]}
  GOARCH=${platform_split[1]}

  os=$GOOS
  if [ $os = "darwin" ]; then
      os="macOS"
  fi

  output_name="goli"
  if [ $os = "windows" ]; then
      output_name+='.exe'
  fi
  echo "Building release/$output_name..."
  mkdir -p release/$latest_version/goliCli
  cp -r resources release/$latest_version/goliCli/
  env GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w" \
    -ldflags "-X github.com/akrabat/rodeo/commands.Version=$latest_version" \
    -o release/$latest_version/goliCli/$output_name
  if [ $? -ne 0 ]; then
      echo 'An error has occurred! Aborting.'
      exit 1
  fi

  # Copy auto update file
  cp autoUpdate.sh release/$latest_version/
  cp autoUpdate.ps1 release/$latest_version/

  zip_name="Goli-${latest_version}-${os}-${GOARCH}"
  pushd release/$latest_version > /dev/null
  echo "${latest_version}" > goliCli/version.txt

  zip -r $zip_name.zip goliCli

  rm -rf goliCli
  popd > /dev/null
done
echo "${latest_version}" > ./release/"${latest_version}"/latest_version.txt
