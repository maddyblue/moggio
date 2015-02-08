#!/bin/bash

set -e

APPNAME=mog
DIR="$APPNAME.app/Contents/MacOS"
OUTPUT="$DIR/$APPNAME"
PA=libportaudio.2.dylib
PALIB=/usr/local/lib/$PA

rm -rf $APPNAME.app
mkdir -p $DIR
go build -o $OUTPUT
chmod +x $OUTPUT
cp $PALIB $DIR/$PA
install_name_tool -change $PALIB @executable_path/$PA $OUTPUT

cat > $DIR/../Info.plist << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
        <key>CFBundleExecutable</key>
        <string>mog</string>
        <key>CFBundleIdentifier</key>
        <string>io.mog.Mog</string>
        <key>CFBundleName</key>
        <string>mog</string>
        <key>CFBundlePackageType</key>
        <string>APPL</string>
        <key>CFBundleSignature</key>
        <string>????</string>
        <key>CFBundleVersion</key>
        <string>0.0.1</string>
</dict>
</plist>
EOF

rm -f mog.zip
zip mog -r mog.app