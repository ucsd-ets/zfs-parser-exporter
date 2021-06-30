# ZFS Parser Exporter

Simple ZFS exporter that relies on parsing the output of `zfs` or `zpool` commands so that you don't have to worry about zfs driver support dependencies.

Currently, only compatible with `zpool iostat` exporter. Works by parsing the output of the `zpool iostat` command so you don't have to worry about go zfs driver support.

## References

- https://www.thegeekdiary.com/how-to-create-virtual-block-device-loop-device-filesystem-in-linux/
- https://docs.oracle.com/cd/E19253-01/819-5461/gaynr/index.html