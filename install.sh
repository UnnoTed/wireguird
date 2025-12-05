if [[ -f "/etc/os-release" ]]; then
    source "/etc/os-release"
    if [[ "${ID}" == "fedora" ]]; then
        echo "not supported yet"
        #sudo rpm
    elif [[ "${ID}" == "ubuntu" || "${ID}" == "debian" || "${ID}" == "linuxmint" ]]; then
        ARCH=${ARCH:-amd64}
        sudo dpkg -i ./build/wireguird_${ARCH}.deb
    elif [[ "${ID}" == "raspbian" ]]; then
        ARCH=${ARCH:-arm64}
        sudo dpkg -i ./build/wireguird_${ARCH}.deb
    fi
fi

