# fireclient
Toy client for firecracker.

This does the equivalent of 

```
curl --unix-socket /tmp/firecracker.sock -i \
    -X PUT 'http://localhost/boot-source'   \
    -H 'Accept: application/json'           \
    -H 'Content-Type: application/json'     \
    -d '{
        "kernel_image_path": "./vmlinux",
        "boot_args": "console=ttyS0 reboot=k panic=1 pci=off"
    }'

curl --unix-socket /tmp/firecracker.sock -i \
    -X PUT 'http://localhost/drives/rootfs' \
    -H 'Accept: application/json'           \
    -H 'Content-Type: application/json'     \
    -d '{
        "drive_id": "rootfs",
        "path_on_host": "./hello-rootfs.ext4",
        "is_root_device": true,
        "is_read_only": false
    }'

curl --unix-socket /tmp/firecracker.sock -i \
     -X PUT "http://localhost/vsocks/root" \
     -H "accept: application/json" \
     -H "Content-Type: application/json" \
     -d "{
            \"id\": \"root\",
            \"guest_cid\": 3
         }"

curl --unix-socket /tmp/firecracker.sock -i \
    -X PUT 'http://localhost/actions'       \
    -H  'Accept: application/json'          \
    -H  'Content-Type: application/json'    \
    -d '{
        "action_type": "InstanceStart"
     }'
```
