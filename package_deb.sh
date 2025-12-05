#!/usr/bin/env sh

echo "wireguird: cleaning..."

ARCH=${ARCH:-amd64}
deb_file="./build/wireguird_${ARCH}.deb"
if [ -e "$deb_file" ]; then
  rm -r "$deb_file"
fi

opt_w_dir="./deb/opt/wireguird/"
if [ -e "$opt_w_dir" ]; then
  rm -r "$opt_w_dir"
fi

mkdir -p "$opt_w_dir"

echo "wireguird: building go binary..."
  go generate
  GOARCH=$ARCH go build -ldflags "-s -w" -trimpath -o "$opt_w_dir""wireguird"

echo "wireguird: copying icons..."
cp -r ./Icon/ "$opt_w_dir"

echo "wireguird: building deb package..."

echo '{"MultipleTunnels":false,"StartOnTray":false,"CheckUpdates":false,"TunnelsPath":"/etc/wireguard","Debug":false}' > "$opt_w_dir""wireguird.settings"

if [ ! -d "./build/" ]; then
  mkdir ./build/
fi

sed -i "s/Architecture: .*/Architecture: $ARCH/" ./deb/DEBIAN/control

dpkg-deb --root-owner-group --build ./deb $deb_file
echo "wireguird: done"
