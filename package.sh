if [[ -f "/etc/os-release" ]]; then
    source "/etc/os-release"
    if [[ "${ID}" == "fedora" ]]; then
        echo "not supported yet"
        #echo "rpm package"
        #./package_rpm.sh
    elif [[ "${ID}" == "ubuntu" || "${ID}" == "debian" || "${ID}" == "linuxmint" || "${ID}" == "raspbian" ]]; then
        echo "deb package"
        ./package_deb.sh
    fi
fi
