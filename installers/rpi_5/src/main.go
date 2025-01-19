// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/siderolabs/go-copy/copy"
	"github.com/siderolabs/talos/pkg/machinery/overlay"
	"github.com/siderolabs/talos/pkg/machinery/overlay/adapter"
)

const (
	board = "rpi5"
)

//go:embed config.txt
var configTxt []byte

func main() {
	adapter.Execute(&rpi5{})
}

type rpi5 struct{}

type rpi5ExtraOptions struct {
	ConfigTxt       string `yaml:"configTxt,omitempty"`
	ConfigTxtAppend string `yaml:"configTxtAppend,omitempty"`
}

func (i *rpi5) GetOptions(extra rpi5ExtraOptions) (overlay.Options, error) {
	return overlay.Options{
		Name: board,
		KernelArgs: []string{
			"console=tty0",
			"console=ttyAMA0,115200",
			"sysctl.kernel.kexec_load_disabled=1",
			"talos.dashboard.disabled=1",
		},
	}, nil
}

func (i *rpi5) Install(options overlay.InstallOptions[rpi5ExtraOptions]) error {
	var f *os.File

	f, err := os.OpenFile(options.InstallDisk, os.O_RDWR|nix.O_CLOEXEC, 0o666)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", options.InstallDisk, err)
	}

	defer f.Close() //nolint:errcheck

	err = copy.Dir(filepath.Join(options.ArtifactsPath, "arm64/firmware/boot"), filepath.Join(options.MountPrefix, "/boot/EFI"))
	if err != nil {
		return err
	}

	err = copy.File(filepath.Join(options.ArtifactsPath, "arm64/rpi5-uefi/RPI_EFI.fd"), filepath.Join(options.MountPrefix, "/boot/EFI/RPI_EFI.fd"))
	if err != nil {
		return err
	}

	err = f.Sync()
	if err != nil {
		return err
	}

	if options.ExtraOptions.ConfigTxt != "" {
		configTxt = []byte(options.ExtraOptions.ConfigTxt)
	}

	configTxt = append(configTxt, []byte("\n"+options.ExtraOptions.ConfigTxtAppend)...)

	return os.WriteFile(filepath.Join(options.MountPrefix, "/boot/EFI/config.txt"), configTxt, 0o644)
}
