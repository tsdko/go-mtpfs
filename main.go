// Copyright 2012 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"os"
	"strings"
	"syscall"
	"time"

	fusefs "github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/hanwen/go-mtpfs/fs"
)

func main() {
	debug := flag.String("debug", "", "comma-separated list of debugging options: usb, data, mtp, fuse")
	usbTimeout := flag.Int("usb-timeout", 5000, "timeout in milliseconds")
	vfat := flag.Bool("vfat", true, "assume removable RAM media uses VFAT, and rewrite names.")
	other := flag.Bool("allow-other", false, "allow other users to access mounted fuse. Default: false.")
	deviceFilter := flag.String("dev", "",
		"regular expression to filter device IDs, "+
			"which are composed of manufacturer/product/serial.")
	storageFilter := flag.String("storage", "", "regular expression to filter storage areas.")
	android := flag.Bool("android", true, "use android extensions if available")
	flag.Parse()

	if len(flag.Args()) != 1 {
		log.Fatalf("Usage: %s [options] MOUNT-POINT\n", os.Args[0])
	}
	mountpoint := flag.Arg(0)

	debugs := map[string]bool{}
	for _, s := range strings.Split(*debug, ",") {
		debugs[s] = true
	}
	mtpOptions := fs.MTPOptions{
		DeviceFilter: *deviceFilter,
		MTPDebug:     debugs["mtp"],
		DataDebug:    debugs["data"],
		USBDebug:     debugs["usb"],
		Timeout:      *usbTimeout,
	}

	opts := fs.DeviceFsOptions{
		RemovableVFat: *vfat,
		Android:       *android,
		StorageFilter: *storageFilter,
		MTPOptions:    mtpOptions,
	}
	root, err := fs.NewDeviceFSRoot(opts)
	if err != nil {
		log.Fatalf("NewDeviceFs failed: %v", err)
	}

	sec := time.Second
	mountOpts := &fusefs.Options{
		MountOptions: fuse.MountOptions{
			SingleThreaded: true,
			AllowOther:     *other,
			Debug:          debugs["fuse"] || debugs["fs"],
		},
		UID:          uint32(syscall.Getuid()),
		GID:          uint32(syscall.Getgid()),
		AttrTimeout:  &sec,
		EntryTimeout: &sec,
	}
	server, err := fusefs.Mount(mountpoint, root, mountOpts)
	if err != nil {
		log.Fatalf("mount failed: %v", err)
	}

	server.WaitMount()
	log.Printf("FUSE mounted")
	server.Wait()
	root.OnUnmount()
}
