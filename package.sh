if [[ -f "/etc/os-release" ]]; then
    source "/etc/os-release"
    if [[ "${ID}" == "fedora" ]]; then
        echo "not supported yet"
        #echo "rpm package"
        #./package_rpm.sh
    elif [[ "${ID}" == "ubuntu" ]]; then
        echo "deb package"
        ./package_deb.sh
	elif [[ "${ID}" == "void" ]]; then
		echo "xbps package"
		./package_xbps.sh
    fi
fi
