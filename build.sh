#!/bin/bash

LIST=(`go tool dist list`)

for item in "${LIST[@]}"; do
    items=(`echo "${item}" | tr "/" "\n"`);
    tags=$(git tag --sort=committerdate);
    target_tag=${tags[0]};
    OS=${items[0]};
    ARCH=${items[1]};
    FOLDER="./dist/brestore-${target_tag}.${OS}-${ARCH}"
    if GOOS="${OS}" GOARCH="${ARCH}" go build -o "$FOLDER/" ./cmd/brestore 2> /dev/null; then
      echo "Generated $OS/$ARCH";
      cp README.md LICENSE "$FOLDER";
      cd dist
      BASENAME="brestore-${target_tag}.${OS}-${ARCH}"
      if [ $OS = "windows" ]; then
          zip -m --quiet -r "${BASENAME}.zip" "${BASENAME}"
        else
          tar --remove-files -czf "${BASENAME}.tar.gz" "${BASENAME}"
      fi
      cd ..
    else
      echo "Skipping $OS/$ARCH";
    fi
done

cd dist

sha256sum *.gz *.zip > sha256sums.txt
