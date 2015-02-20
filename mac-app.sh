#!/bin/bash

set -e

APPNAME=mog
DIR="$APPNAME.app/Contents/MacOS"
OUTPUT="$DIR/$APPNAME"
PA=libportaudio.2.dylib
PALIB=/usr/local/lib/$PA
DARWIN386=/usr/local/go-darwin-386

rm -rf $APPNAME.app
mkdir -p $DIR

go build -o $OUTPUT-amd64 -ldflags "-linkmode=external"
GOROOT=$DARWIN386 $DARWIN386/bin/go build -o $OUTPUT-386 -ldflags "-linkmode=external"

lipo $OUTPUT-* -output $OUTPUT -create
rm $OUTPUT-*
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
</dict>
</plist>
EOF

rm -f mog.zip
zip mog -r mog.app
