if [[ -f "/etc/os-release" ]]; then
    source "/etc/os-release"
    if [[ "${ID}" == "fedora" ]]; then
        echo "not supported yet"
        #sudo rpm
    elif [[ "${ID}" == "ubuntu" ]]; then
        ARCH=${ARCH:-amd64}
        sudo dpkg -i ./build/wireguird_${ARCH}.deb
    fi
fi

