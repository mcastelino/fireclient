package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/go-openapi/strfmt"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/mcastelino/fireclient/client"
	models "github.com/mcastelino/fireclient/client/models"
	ops "github.com/mcastelino/fireclient/client/operations"
)

func NewFireClient(socketPath string) *client.Firecracker {
	httpClient := client.NewHTTPClient(strfmt.NewFormats())

	socketTransport := &http.Transport{
		DialContext: func(ctx context.Context, network, path string) (net.Conn, error) {
			addr, err := net.ResolveUnixAddr("unix", socketPath)
			if err != nil {
				return nil, err
			}

			return net.DialUnix("unix", nil, addr)
		},
	}

	transport := httptransport.New(client.DefaultHost, client.DefaultBasePath, client.DefaultSchemes)
	transport.Transport = socketTransport

	httpClient.SetTransport(transport)

	return httpClient
}

func getInstanceState(fireClient *client.Firecracker) {
	{
		resp, err := fireClient.Operations.DescribeInstance(nil)
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
		fmt.Printf("%#v\n", resp.Payload)
	}

	{
		resp, err := fireClient.Operations.GetMachineConfig(nil)
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
		fmt.Printf("%#v\n", resp.Payload)
	}

	{
		resp, err := fireClient.Operations.GetMmds(nil)
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
		fmt.Printf("%#v\n", resp.Payload)
	}
}

func main() {

	fireClient := NewFireClient("/tmp/firecracker.sock")

	if fireClient == nil {
		os.Exit(1)
	}

	getInstanceState(fireClient)

	/*
		curl --unix-socket /tmp/firecracker.sock -i \
		    -X PUT 'http://localhost/boot-source'   \
		    -H 'Accept: application/json'           \
		    -H 'Content-Type: application/json'     \
		    -d '{
		        "kernel_image_path": "./vmlinux",
		        "boot_args": "console=ttyS0 reboot=k panic=1 pci=off"
		    }'
	*/
	{
		kernel := "./vmlinux"
		cmdLine := "init=/usr/lib/systemd/systemd systemd.unit=kata-containers.target systemd.mask=systemd-networkd.service systemd.mask=systemd-networkd.socket systemd.mask=systemd-journald.service systemd.mask=systemd-journald.socket systemd.mask=systemd-journal-flush.service systemd.mask=systemd-udevd.service systemd.mask=systemd-udevd.socket systemd.mask=systemd-udev-trigger.service systemd.mask=systemd-timesyncd.service systemd.mask=systemd-update-utmp.service systemd.mask=systemd-tmpfiles-setup.service systemd.mask=systemd-tmpfiles-cleanup.service systemd.mask=systemd-tmpfiles-cleanup.timer systemd.mask=tmp.mount systemd.mask=systemd-random-seed.service agent.log=debug console=ttyS0 reboot=k panic=1 pci=off root=/dev/vda1 random.trust_cpu=on"
		bootSrcParams := ops.NewPutGuestBootSourceParams()
		src := &models.BootSource{
			KernelImagePath: &kernel,
			BootArgs:        cmdLine,
		}

		bootSrcParams.SetBody(src)

		_, err := fireClient.Operations.PutGuestBootSource(bootSrcParams)
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
	}

	/*
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
	*/
	{
		driveID := "rootfs"
		hostPath := "./kata.img"
		driveParams := ops.NewPutGuestDriveByIDParams()
		driveParams.SetDriveID(driveID)
		isReadOnly := false
		isRootDevice := false
		drive := &models.Drive{
			DriveID:      &driveID,
			IsReadOnly:   &isReadOnly,
			IsRootDevice: &isRootDevice,
			PathOnHost:   &hostPath,
		}
		driveParams.SetBody(drive)
		_, err := fireClient.Operations.PutGuestDriveByID(driveParams)
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
	}

	{
		cfg := ops.NewPutGuestNetworkInterfaceByIDParams()
		ifaceID := "tap0"
		ifaceCfg := &models.NetworkInterface{
			AllowMmdsRequests: false,
			GuestMac:          "02:ca:fe:ca:fe:01",
			IfaceID:           &ifaceID,
			HostDevName:       &ifaceID,
		}
		cfg.SetBody(ifaceCfg)
		cfg.SetIfaceID(ifaceID)
		_, err := fireClient.Operations.PutGuestNetworkInterfaceByID(cfg)
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
	}

	{
		param := ops.NewPutMachineConfigurationParams()
		cfg := &models.MachineConfiguration{
			MemSizeMib: 512,
			VcpuCount:  4,
		}
		param.SetBody(cfg)
		_, err := fireClient.Operations.PutMachineConfiguration(param)
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
	}

	/*
		curl --unix-socket /tmp/firecracker.sock -i \
		     -X PUT "http://localhost/vsocks/root" \
		     -H "accept: application/json" \
		     -H "Content-Type: application/json" \
		     -d "{
		            \"id\": \"root\",
		            \"guest_cid\": 3
		         }"
	*/
	{
		vsockParams := ops.NewPutGuestVsockByIDParams()
		vsockID := "root"
		vsock := &models.Vsock{
			GuestCid: 3,
			ID:       &vsockID,
		}
		vsockParams.SetID(vsockID)
		vsockParams.SetBody(vsock)
		_, _, err := fireClient.Operations.PutGuestVsockByID(vsockParams)
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
	}

	/*
		curl --unix-socket /tmp/firecracker.sock -i \
		    -X PUT 'http://localhost/actions'       \
		    -H  'Accept: application/json'          \
		    -H  'Content-Type: application/json'    \
		    -d '{
		        "action_type": "InstanceStart"
		     }'
	*/

	/* This is where the VM is actually launched */
	{
		actionParams := ops.NewCreateSyncActionParams()
		actionInfo := &models.InstanceActionInfo{
			ActionType: "InstanceStart",
			//Payload:    "",
		}
		actionParams.SetInfo(actionInfo)
		_, err := fireClient.Operations.CreateSyncAction(actionParams)
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
	}

	time.Sleep(250 * time.Millisecond)
	getInstanceState(fireClient)
}
