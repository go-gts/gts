#!/bin/sh

set -ex

mkdir -p "$PREFIX/bin"
mkdir -p "$PREFIX/share/man/man1"
mkdir -p "$PREFIX/share/man/man7"

cp "$SRC_DIR/gts" "$PREFIX/bin"
cp "$SRC_DIR/togo" "$PREFIX/bin"

chmod +x "$PREFIX/bin/gts"
chmod +x "$PREFIX/bin/togo"

for FILE in "$SRC_DIR"/man/*.1; do
    cp "$FILE" "$PREFIX/share/man/man1"
done

for FILE in "$SRC_DIR"/man/*.7; do
    cp "$FILE" "$PREFIX/share/man/man7"
done
