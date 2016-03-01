#!/bin/bash

set -e

APPNAME=moggio
DIR="$APPNAME.app/Contents/MacOS"
OUTPUT="$DIR/$APPNAME"
PA=libportaudio.2.dylib
PALIB=/usr/local/lib/$PA

rm -rf $APPNAME.app
mkdir -p $DIR
go build -o $OUTPUT -ldflags "-linkmode=external"
chmod +x $OUTPUT
cp $PALIB $DIR/$PA
install_name_tool -change $PALIB @executable_path/$PA $OUTPUT

cat > $DIR/../Info.plist << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
        <key>CFBundleExecutable</key>
        <string>moggio</string>
        <key>CFBundleIdentifier</key>
        <string>io.moggio.Moggio</string>
        <key>CFBundleName</key>
        <string>moggio</string>
        <key>CFBundlePackageType</key>
        <string>APPL</string>
        <key>CFBundleSignature</key>
        <string>????</string>
</dict>
</plist>
EOF

rm -f moggio.zip
zip moggio -r moggio.app
