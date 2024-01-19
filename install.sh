if [[ -f "/etc/os-release" ]]; then
    source "/etc/os-release"
    if [[ "${ID}" == "fedora" ]]; then
        echo "not supported yet"
        #sudo rpm
    elif [[ "${ID}" == "ubuntu" ]]; then
        sudo dpkg -i ./build/wireguird_amd64.deb
	elif [[ "${ID}" == "void" ]]; then
		xbps-rindex -a *.xbps
		sudo xbps-install --repository=$PWD wireguird
    fi
fi

