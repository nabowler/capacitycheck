# Capacity Check

A simple library and CLI to help verify the advertised capacity of storage media.

This program will write a single file of pseudo-random data to the requested directory. A checksum of the data is calculated while creating the file. Once the file has been fully written and synched to disk, the file is read, and the checksum of the read file is compared to the expected checksum. The file is then deleted.

If the checksums do not agree, the media _may_ be lying about its advertised capacity.

## To build

```shell
go build cmd/capacitycheck/capacitycheck.go
```

## Usage

```shell
# write a 63GB file to /media/foo/bar with a read/write buffer of 1MB
./capacitycheck -d /media/foo/bar -s 63GB -b 1MB
```
