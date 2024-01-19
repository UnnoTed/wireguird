echo "wireguird: cleaning..."

opt_w_dir="./deb/opt/wireguird/"
if [ -e "$opt_w_dir" ]; then
  rm -r "$opt_w_dir"
fi

mkdir -p "$opt_w_dir"

echo "wireguird: building go binary..."
time {
  go generate
  go build -ldflags "-s -w" -trimpath -o "$opt_w_dir""wireguird" -p $(nproc) -v -x
}

echo "wireguird: copying icons..."
cp -r ./Icon/ "$opt_w_dir"

echo "wireguird: building xbps package..."

touch "$opt_w_dir""wireguird.settings"

if [ ! -d "./build/" ]; then
  mkdir ./build/
fi

xbps-create \
	-A noarch \
	-n wireguird-1.0_1 \
	-s "wireguard gtk gui for linux" \
	./deb

echo "wireguird: done"
