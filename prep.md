## Blank Image

1. `dd if=/dev/zero of=blank_img bs=1M count=1500`
2. `losetup --find --show blank_img`
3. Follow the `fdisk` part of https://archlinuxarm.org/platforms/armv6/raspberry-pi

   ```
   At the fdisk prompt, delete old partitions and create a new one:
   Type o. This will clear out any partitions on the drive.
   Type p to list partitions. There should be no partitions left.
   Type n, then p for primary, 1 for the first partition on the drive, press ENTER to accept the default first sector, then type +200M for the last sector.
   Type t, then c to set the first partition to type W95 FAT32 (LBA).
   Type n, then p for primary, 2 for the second partition on the drive, and then press ENTER twice to accept the default first and last sector.
   Write the partition table and exit by typing w.
   ```
4. `partprobe /dev/loop0`
5. `mkfs.vfat /dev/loop0p1`
6. `mkfs.xfs -m reflink=1 /dev/loop0p2`
7. Double compress the image: `xz < blank_img | xz > blank_img.xz.xz`

## OS

```
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
