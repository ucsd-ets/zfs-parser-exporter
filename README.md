# ZFS Exporter

Simple ZFS exporter that relies on parsing the output of `zfs` or `zpool` commands so that you don't have to worry about zfs driver support dependencies.

Currently, only compatible with `zpool iostat` exporter. Works by parsing the output of the `zpool iostat` command so you don't have to worry about go zfs driver support.

