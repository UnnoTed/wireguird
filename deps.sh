if [[ -f "/etc/os-release" ]]; then
    source "/etc/os-release"
    if [[ "${ID}" == "fedora" ]]; then
        echo "not supported yet"
        #sudo dnf install wireguard-tools gtk3-devel golang resolvconf
    elif [[ "${ID}" == "ubuntu" || "${ID}" == "debian" || "${ID}" == "linuxmint" || "${ID}" == "raspbian" ]]; then
        sudo apt install wireguard-tools libgtk-3-dev libayatana-appindicator3-dev resolvconf dpkg-dev
    fi
fi
