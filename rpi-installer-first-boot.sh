#!/bin/bash
set -o pipefail -e -u

_tmp_path=/tmp/rpi-installer-first-boot
mkdir -p "${_tmp_path}"
cd "${_tmp_path}"

# Extract installer config files
_arch_img_size=1572864000
tail -c +$(( 1 + _arch_img_size )) /dev/mmcblk0 | tar -x # tar ignores garbage at the end

# Grow the root partition
echo ", +" | sfdisk --force -N 2 /dev/mmcblk0
partx -u /dev/mmcblk0
xfs_grow -d /

# Set hostname
hostnamectl set-hostname "$(cat hostname)"

# Set timezone TODO

# Set arch mirror TODO

# Add authorized SSH keys TODO

# Configure WiFi
_wpa_supp_path=/etc/wpa_supplicant/wpa_supplicant-wlan0.conf
if [ -f wifi_ssid ]; then

  cat <<EOF > /etc/systemd/network/wlan0.network
[Match]
Name=wlan0

[Network]
DHCP=ipv4
EOF

  cat <<EOF > "${_wpa_supp_path}"
ap_scan=1
fast_reauth=1
country=$(cat wifi_country)
network={
 ssid="$(cat wifi_ssid)"
EOF

  if [ -f wifi_psk ]; then
    echo "  psk=$(cat wifi_psk)" >> "${_wpa_supp_path}"
  else
    echo "  key_mgmt=NONE" >> "${_wpa_supp_path}"
  fi
  echo "}" >> "${_wpa_supp_path}"
  systemctl reload systemd-networkd.service
  systemctl enable --now wpa_supplicant@wlan0.service
fi

# Cleanup
rm -rf "${_tmp_path}"
rm -f /etc/systemd/system/{multi-user.target/,}rpi-installer-first-boot.service /opt/rpi-installer-first-boot.sh
