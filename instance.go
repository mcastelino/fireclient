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
		cmdLine := "console=ttyS0 reboot=k panic=1 pci=off"
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
		hostPath := "./hello-rootfs.ext4"
		driveParams := ops.NewPutGuestDriveByIDParams()
		driveParams.SetDriveID(driveID)
		isReadOnly := false
		isRootDevice := true
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
		driveID := "hotDummy"
		hostPath := "./dummy.img"
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
			HostDevName:       "tap0",
			State:             "Attached",
		}
		cfg.SetBody(ifaceCfg)
		cfg.SetIfaceID(ifaceID)
		_, err := fireClient.Operations.PutGuestNetworkInterfaceByID(cfg)
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

	/* Make sure it really boots and sees the drive */
	time.Sleep(5000 * time.Millisecond)
	getInstanceState(fireClient)

	/* Now update the driver to point to a new one and rescan */
	{
		driveID := "hotDummy"
		hostPath := "./real.img"
		driveParams := ops.NewPatchGuestDriveByIDParams()
		driveParams.SetDriveID(driveID)
		drive := &models.PartialDrive{
			DriveID:    &driveID,
			PathOnHost: &hostPath, //This is the only property that can be modified on the fly
		}
		driveParams.SetBody(drive)
		_, err := fireClient.Operations.PatchGuestDriveByID(driveParams)
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
	}

	{
		actionParams := ops.NewCreateSyncActionParams()
		actionInfo := &models.InstanceActionInfo{
			ActionType: "BlockDeviceRescan",
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
