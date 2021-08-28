## Blank Image

1. Follow the `fdisk` part of https://archlinuxarm.org/platforms/armv6/raspberry-pi
2. Instead of `ext4`, run `mkfs.xfs -m reflink=1 /dev/....`
3. Update the code with UUIDs of the two partitions.
4. Compress the image: `xz blank_img`

## OS

```
cd /tmp
xz -d < blank_img.xz > blank_img
losetup --find --show blank_img
mkdir root boot
mount /dev/loop0p1 boot
mount /dev/loop0p2 root
# TODO verify integrity of the tar
curl -sL 'http://os.archlinuxarm.org/os/ArchLinuxARM-rpi-latest.tar.gz' -o pi.tar.gz
tar xf pi.tar.gz --exclude='boot/**' -C root
tar xf pi.tar.gz ./boot -C boot

_arch() {
  docker run --rm -it \
    -v $(pwd)/root:/piroot \
    -v $(pwd)/root/etc/pacman.d/mirrorlist:/etc/pacman.d/mirrorlist \
    -v $(pwd)/root/usr/share/pacman/keyrings:/usr/share/pacman/keyrings \
    --mount type=tmpfs,destination=/piroot/var/lib/pacman/sync archlinux "${@}"
}

_arch pacman-key --config /piroot/etc/pacman.conf --gpgdir /piroot/etc/pacman.d/gnupg/ --init 
_arch pacman-key --config /piroot/etc/pacman.conf --gpgdir /piroot/etc/pacman.d/gnupg/ --populate archlinuxarm
_arch pacman -r /piroot --arch armv6h --gpgdir /piroot/etc/pacman.d/gnupg/ --config /piroot/etc/pacman.conf --cachedir /tmp --dbpath /piroot/var/lib/pacman/ -Sy
_arch pacman -r /piroot --arch armv6h --gpgdir /piroot/etc/pacman.d/gnupg/ --config /piroot/etc/pacman.conf --cachedir /tmp --dbpath /piroot/var/lib/pacman/ -S xfsprogs --noconfirm

umount root boot
rmdir root boot
losetup --detach /dev/loop0
xz < blank_img > arch_img.xz
```
