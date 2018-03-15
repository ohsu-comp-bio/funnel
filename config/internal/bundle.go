// Code generated by go-bindata.
// sources:
// config/default-config.yaml
// config/gridengine-template.txt
// config/htcondor-template.txt
// config/pbs-template.txt
// config/slurm-template.txt
// DO NOT EDIT!

package config

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _configDefaultConfigYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xec\x5a\x51\x73\x1b\x39\x8e\x7e\xef\x5f\x81\xb3\xb2\x75\x49\x95\x24\xcb\x9b\x9a\xab\x1d\x55\xf9\xc1\x96\x3d\x8e\x2f\x8e\xe3\xb3\x94\xcb\xdd\x93\x8b\x6a\xa2\xd5\x1c\xb3\xc9\x1e\x92\x6d\x59\xf1\xe5\xbf\x5f\x01\x64\xb7\x24\xc7\x8e\x3d\xbb\xce\x56\x1e\x56\x4f\x12\x1b\x04\x40\x00\x1f\x08\xa0\xd5\x83\x59\x89\x60\x44\x85\x60\x0b\x08\x25\x82\xc8\x83\xba\x41\xf0\xe8\x6e\xd0\x81\x14\x41\xcc\x85\x47\x98\x8b\xfc\x1a\x8d\xcc\x7a\x70\x70\x23\x94\x16\x73\xdd\xad\xf9\x31\xcc\xad\x0e\x72\xde\x07\xb9\x32\xa2\xb2\xf4\x0d\xb5\xf0\x41\xe5\x7d\xa8\xac\x59\x58\x39\xcf\x8e\x12\xa7\x96\x38\xcb\x1e\x95\x9d\xdb\xaa\x6e\xc2\x53\x32\xb5\xcd\x85\xee\x43\x19\x72\x6b\xa4\x75\x7d\xf0\xba\x71\x55\x1f\xea\xb9\xef\xc3\xc2\x29\x89\x66\xa1\x0c\xf6\xa1\x12\xa6\x21\x4a\xb1\xf4\x83\xb9\x08\x79\x99\x4d\xa2\x80\xc4\xe3\x3b\x9a\xe0\x0d\x9a\x00\x4b\xa7\x02\xba\x56\xf4\x6b\xff\x66\xf8\xa8\x4a\x8b\xfe\x33\x6c\xd1\x87\x6b\x51\x5c\x8b\xec\x98\xb8\x7f\x66\xe6\x7e\x0c\x19\xc0\xa0\xb5\x0d\x7d\xd5\x76\x91\x65\x67\x76\xb1\x40\x37\xce\x00\x7a\x40\xdf\x95\x59\x80\xc6\x1b\xd4\x7e\x0c\x12\xe7\xcd\xa2\x0f\xca\x14\xb6\x0f\xe8\x9c\x75\x19\xc0\x19\x3d\x1c\xf3\x22\x6f\x62\xf6\xc4\xcb\x43\xb0\x10\x4a\xe5\xa1\x16\xa1\x1c\xc2\x69\x01\x58\xd5\x61\xd5\x8f\x0f\x85\x43\x3e\x67\x40\x43\x84\x3e\x48\x74\x6e\x98\x01\x7c\x6c\x42\xdd\x84\xdf\x94\xc6\x31\xec\xec\x64\xd9\x94\x03\x23\x6a\xf4\xce\xfa\xb0\x69\xb5\xdf\x1a\x63\x50\xa7\xd8\xa1\xcd\x44\x70\x2e\xaa\xd6\xd2\xa5\xf5\x21\xe3\x9d\x17\xd6\x05\x68\x3c\x4a\x28\xac\x83\x77\xb3\xd9\x05\x79\xbd\x6a\x8c\xca\x45\x50\xd6\x80\x30\x92\x59\x2e\x71\x0e\x52\xf8\x72\x6e\x85\x93\xcc\x72\x36\xbb\xa0\xdd\x63\xf8\xdb\x68\x34\x7a\x88\xdb\xe5\xc5\x64\x9b\x19\x6d\xbb\xbc\x98\xc4\x5d\xbf\x8e\x7e\x4d\xbb\x2e\xf1\x8f\x46\x39\xf2\x9f\x57\x39\x88\x26\x94\x68\x42\x2b\x9f\x18\x91\xfc\x84\x83\x83\x8b\x53\x0f\x8d\x27\xf3\x0b\xa8\x85\xf7\x4b\x1b\xd5\xe9\x91\x21\x49\x34\xc5\xd9\x35\x82\x6f\x1c\x92\x01\x6b\x67\x6b\x74\x7a\x05\x0e\x7d\x70\x2a\x0f\x20\xf2\x1c\x7d\xf2\x02\xc5\xb8\x29\xd4\x02\x0a\xa5\x91\xb9\xbc\xc6\xe1\x62\x08\x79\x59\x59\x09\xff\x31\x1a\x41\xc1\xa6\x1c\x46\xb2\xe1\xaa\xd2\x6f\xe2\x49\x93\xe8\x31\x88\x79\xbe\xf7\xd7\xb7\xf1\x24\xa7\x26\xd7\x8d\x44\x10\xb0\x33\x11\x79\x89\x83\x89\x35\xc1\x59\x3d\x06\x63\x07\x3e\x58\x87\x3b\xd1\xc6\x25\x0a\x89\x0e\x94\x81\x13\x0c\xbb\x67\xca\x07\xd2\xaf\xb6\xc6\xa3\x67\x4e\xac\x79\x8c\xfa\x5c\xe4\x25\x9d\x77\xbe\x02\x65\x02\xba\x0a\xa5\x12\x6e\xc5\x16\x51\x39\x7a\x3a\xfd\x91\xf2\x04\x01\xe2\xcd\x82\xc7\x10\x5c\x83\xc9\xbc\xe4\x06\xad\x98\x95\x35\x06\x73\xb6\x6b\x50\x15\xda\x26\x24\xd3\x19\x30\xc2\x58\x8f\x04\x61\x9f\xdc\x34\xe1\x3d\xb3\x48\x37\x26\x63\xb4\x1f\xe8\xc1\xdb\x11\x24\xea\x28\x84\x70\x5b\x89\x5b\x55\x35\x15\x98\xa6\x9a\xa3\xe3\x58\x54\x15\x7a\x08\xa5\x08\x20\xc0\xe1\x1f\x0d\xfa\x00\x4b\xa5\x35\xcc\x11\x1c\x06\xa7\x52\xa8\x14\x42\xe9\xc6\xc5\xb3\xf4\x80\x64\xc2\x1c\xc3\x12\xd1\x24\x32\x0f\x85\xd5\xda\x2e\x3d\x08\x03\x78\x5b\x5b\x43\x31\x22\x34\x83\xde\x16\x05\xf8\x20\x5c\xe0\xb0\x08\xf0\x4b\xa7\x1b\x71\x6b\x6a\xb2\xe6\x1e\x54\xca\x34\x01\x37\xcf\xf6\x41\xdc\x5e\x46\xee\x63\xd8\x1b\x65\x6d\xfe\xf1\x79\x89\xb2\xd1\xe4\x1f\xbf\x8e\x66\x0a\x96\x0f\x9c\xc1\xee\xe7\xc5\x21\x64\xd3\x76\x4b\x8b\xc7\x25\xd8\x22\x41\xd8\x35\x06\xc4\x26\xd3\x80\xae\x83\xc3\x43\xb6\x6f\x99\x5d\x0a\x4a\x8d\x7b\x9b\x66\xdf\x4b\x27\xeb\xa4\x54\xc2\xac\x20\x08\x7f\xcd\x01\xdd\x0a\xa1\xc0\xb2\x06\xb7\x45\xb5\x6c\x27\x65\x63\xae\xf9\xc0\x2d\x13\x6d\xcd\x82\xb6\x2f\x85\x0a\x9d\xdd\x9b\x5a\x8a\x80\x1e\xe6\x58\x58\x47\xce\x75\xd7\x11\x75\xc6\x4a\x04\x89\x42\x3e\xa6\xff\xb9\x95\x78\xa1\xcc\xe2\x91\xd0\xd9\xf0\xc4\x03\xe2\xc9\xd4\x49\x06\xa7\x3f\xe1\x42\xff\xbe\x0e\xe4\x8a\x67\x69\x71\x6a\xd4\x3a\x80\xdf\x8e\xb6\xd4\xf8\x25\xa9\xe1\xb3\x8c\x48\xc7\x6d\x0e\x49\xc9\x38\xe9\x70\x7a\xd4\xc5\xab\x68\x82\xad\x04\x25\x26\xad\x57\xb0\x40\x43\xb6\x45\x96\x7f\x7a\x14\x73\x72\x62\xd1\xe9\x57\x0a\xb2\x1f\x1a\x50\x52\x23\x1f\x8d\xce\x8a\x14\x4c\xc2\x30\x59\x82\x61\x1f\x54\x02\x86\x2f\x9b\x00\xd2\x2e\x53\x74\x0c\xf6\xa0\x42\x61\x08\x44\xe8\x90\x02\xd2\xd8\x0e\xbb\x30\x6a\x1f\xc6\x05\x50\x15\x27\x87\x80\x7a\x05\xa2\x08\x18\xa3\xb6\x50\xce\x07\x0e\x12\xe2\xd9\xd9\x63\xb0\x17\x15\x3e\x60\x53\x45\xe9\xdb\x67\x0c\x6e\x45\x4e\x90\x18\x30\x0f\xb0\x24\x18\x3b\xf4\xb6\x71\x39\xc6\x8b\x4a\x74\x77\x6f\xb0\xa0\xc2\x10\x98\xe1\x11\x16\xca\x90\x9f\x2e\x3b\x62\x15\x4f\xcb\x82\x62\x2a\x6d\x62\x60\x82\xbd\x41\x47\x35\x82\x8f\x77\xe2\x1c\x4b\x71\xa3\x2c\x5f\x5a\xdd\x76\xf2\x0d\x31\x9e\x5c\x7c\xf2\x6b\x99\xc3\x76\xb5\x6e\xfc\x18\xf8\x2e\xe1\x74\x77\xf0\x61\x4d\xd3\xe7\x14\x7b\xd8\x92\x5e\x8a\xea\x64\x3e\x86\xd1\xb0\xa3\x3e\x52\xfe\x1a\x7c\x2d\x72\x7c\x74\x13\x91\x6c\xec\xea\xc1\x6f\xec\xc7\xe5\x80\xef\x7f\x08\x0d\x9d\x75\xf8\x2d\xee\xfd\xca\xe4\xb0\x54\xa1\x7c\xf8\x4a\x7e\x28\x66\x3f\x31\xe6\x22\xee\x7f\xd9\x0e\xd6\x2e\xdb\x7e\xb6\xee\xba\xcd\x33\x54\x0d\x78\xc8\x1d\x52\x20\x82\x6c\x1c\x59\xbd\x76\x96\x2e\x38\xfa\xda\x86\x6e\x5b\x50\xb0\x1b\x94\x07\xa9\x1c\xe6\xc1\xba\x15\x09\x25\x86\x47\xca\x8d\x61\xb8\x1b\x2f\xbb\xc1\xd2\xba\xeb\x81\x54\xee\x4f\x1d\xb7\xb6\x5a\x73\x88\xe7\xc2\xe4\x74\x52\xb5\x30\x42\xfb\x47\x4e\x7a\x61\xb5\x56\x66\xf1\xfd\xa3\xfe\x19\x63\xa3\x91\x54\x2c\xd9\x26\xec\xa2\x73\x1c\xed\x54\x50\x75\x69\x2c\x5d\xf7\x0f\xb8\x61\x8a\x21\xc4\xac\xa2\x98\x6c\x14\xcd\xe6\xd0\x37\x3a\xa4\xc8\xf5\x84\x22\xd4\x92\x02\x94\x68\x23\x57\x49\x79\x5d\x99\x85\x8e\x38\x66\x6e\x6b\xd8\xe1\x2d\xe6\x4d\xb0\x0e\xf0\x56\x05\xff\x98\xcb\xcf\xec\xe2\x39\x5e\xa7\xcd\x1f\xc4\x2d\xcc\x57\xe9\x30\x5c\x4d\xb0\xbd\x37\x4e\x9d\x60\xd6\x1e\x3e\xf1\x9f\x09\xa5\xa7\xea\x4b\x7b\x97\x50\x0a\x1e\xc1\xfb\xc3\xc8\xf4\xdc\xba\x2a\x82\x9d\x8a\x3c\x8e\x2d\x90\xa8\x91\xc4\xa8\xe0\x79\x89\x4e\xdc\x85\x4c\x3a\x61\x3c\x5d\xe7\x8c\x19\x19\xcf\xd6\x0c\x69\x19\x8b\x91\x74\xf1\x6f\x22\xfa\x0c\xc5\x0d\x76\xf1\x56\x08\xed\x31\xcb\x7a\x83\x97\xfd\x64\x3d\x68\x7b\x1c\xaa\x16\xe4\xae\x75\xc0\x15\x3e\xa4\x12\x7f\xf7\x9d\x30\x52\xa3\xf3\x2f\x2f\x3a\x3b\xb4\x3a\x1c\x1d\x8e\x53\x8d\x48\xd8\x8f\x71\xd7\xf5\x6f\xa9\xd0\xa4\x67\x0f\x20\x2e\xfd\x1e\x52\x5b\x76\xc4\x7d\x4b\xcb\xec\x50\x78\xe4\x12\x3f\x58\xaa\x49\xd8\xf3\x6d\x67\x03\x81\xed\x4d\xc9\x9d\xbe\xb4\xa4\xe3\x54\xbc\xc6\x2c\xff\x79\x0a\x0e\x17\xca\x1a\xce\xac\xf4\x85\xef\xac\xf6\xd9\x41\xac\x8a\xaf\x71\x05\xa7\x47\x19\xc0\x7b\x5c\x6d\x3d\x9f\x62\xee\x30\xb4\x64\xef\x71\x45\x15\x05\xaf\xc5\xab\xef\x38\xf6\x56\xe9\xe4\x0e\x0b\x75\xbb\xa9\xaa\x32\x12\x6f\xd1\xc3\x6b\x8a\xcd\x7e\xec\xe7\x7c\x9f\x6f\x49\x4f\x15\xf5\x29\x3d\x8f\xdb\xb6\xd4\xfe\x74\x79\xd6\x36\x35\xa9\x7b\xf3\x28\x5c\x5e\x6e\x20\xf8\xd3\xe5\xd9\x18\xca\x10\xea\xf1\xee\x6e\xd7\xdd\x8c\x7f\xfd\x2b\x35\x25\x3d\x38\xb1\x96\xf0\x39\xd1\xb6\x91\x1c\x17\x11\x38\x0c\x91\xd6\x29\xc3\xac\x7b\x40\xfa\x5f\x38\xfb\x3b\xe6\xa1\x3b\x7e\xeb\x47\x91\xe7\xb6\xa1\x2a\xda\xa1\x8c\xd5\xa7\x67\x77\x46\x04\x7c\xe4\xe0\x17\x9a\x3b\xba\xda\x7a\xaf\xf8\x2a\xd9\x24\x7e\xb8\x92\x90\xca\xe7\x74\x0b\x62\xac\xea\x0a\x67\xab\x78\x5e\x73\xa3\x9c\x35\x15\x1a\xae\xd2\x27\x6b\x46\x5d\x13\x08\x90\x7d\xa0\x56\xb6\x0d\x92\x03\x29\x9d\x87\xd2\x52\xa2\xe2\xee\x59\x4a\x87\xde\x73\xf5\xdc\xb6\x51\x28\x93\xed\x38\xfd\xf0\x8e\x78\xbf\x0e\x36\x7a\x43\xbe\xf7\xda\x90\x55\x7e\x3b\x84\x39\x0c\xb9\x2e\xa6\xdb\x4d\x19\x48\x3a\x6c\xa4\xa5\x98\x65\x69\x07\x77\x27\xdd\xc8\x61\xc3\xb3\xb3\xb6\x66\x49\xaa\x56\x6c\xdb\xd4\x34\xdc\x2b\x07\x53\xf3\x47\xc5\x34\x77\x49\x12\x96\x25\x9a\x68\x2e\x2e\x6e\xda\xc6\x86\x0a\x54\x23\x81\xfb\x46\x6a\x17\xa8\xf8\xa7\xfe\x8f\x6b\x8d\xae\x0c\xf1\x74\x3d\x5a\x43\x9e\x8a\xcd\xd6\x5a\x95\x2f\xe8\x6c\x3f\x36\x86\x42\x6b\xa8\xc4\x0a\xe6\xda\xe6\xd7\xa4\x08\x92\x0e\xa4\x15\x89\x89\x8a\xad\x1b\xaa\xb6\x6b\x9b\x23\xa0\x27\x3c\x2a\x5f\xc6\xe2\xf0\xe9\x02\x94\x03\xdd\xa3\x63\xc3\x92\xfe\x6d\x77\xc9\xe3\x03\x17\xc3\x61\x2b\xea\x92\x37\x95\x51\xdc\x04\x6d\xf7\xcc\xcc\x4f\x52\xfd\x6f\xcd\xb6\xe7\x24\x55\x66\x28\xa9\x95\xa4\xf5\xa3\x75\x52\x42\xcd\xba\xb6\x5a\xa4\xe8\x5a\x77\xb9\x04\xf1\xf7\xa2\xb8\x16\x63\xc6\x3d\xc7\x4f\x1b\x36\x4c\x3a\xb3\xb5\xca\x3b\x07\xff\x88\xa4\x9e\x86\x45\x70\x98\xc6\x3c\x3f\x20\x7b\xbf\x9b\x4d\x78\x86\x15\xd1\x34\x6b\x9c\x01\xea\x2d\x39\x59\xf8\x20\x02\x35\xad\xb9\x35\xb9\xd2\xe8\x86\xf0\xb9\x44\x03\x68\x28\xe5\xca\x7e\x5b\x59\xac\x07\x1e\xe8\xd7\xd5\xdf\xbb\x8b\x09\xb3\x5c\x77\x81\xc1\x42\xa1\x8c\x6c\x7b\x37\x6e\x91\x1d\x82\x0f\x4d\x7e\x4d\x71\x2a\xe0\x8f\x06\x1b\x02\x2b\xcb\xa5\x32\xc2\x39\xeb\xa8\xe6\x48\xed\x5f\x57\xd9\xb4\x37\x7e\xa4\xa4\x2c\xe5\x24\x55\x25\xab\x8d\xc9\xc0\x65\xa7\x77\x1a\x0d\xc4\xc1\x4b\x5a\xa4\xda\x83\xa2\xbf\x5c\x97\x54\xe5\x37\xe3\x3f\xfe\x2d\x1c\xfa\x28\x88\xd1\x14\x0f\xfd\xef\xbe\x1b\x11\x26\x14\x84\xd2\x7a\x32\x56\x6d\x5d\x58\xc7\xdb\x9a\x68\x4b\xf2\x18\xf6\xfe\xb6\x0d\x0d\x78\x3b\xda\x00\xc7\x0c\xab\x5a\x33\xdd\xff\x71\xc4\x35\x46\x51\xf8\x21\xec\xc3\x8d\x30\x4a\x6b\xc1\xcb\x0b\x0c\x68\x6e\x60\x1f\x66\xf1\x7c\x90\x6a\x14\xee\x4f\xf6\xe1\xee\x6e\x78\xdc\xfd\xfe\xfa\x95\x09\x84\x5b\x34\x94\x5f\x3d\xec\xb7\xb5\x0f\x75\xeb\x83\x41\x9a\x0a\xdd\xdd\x0d\x27\xfc\xed\xeb\x57\x18\x0c\xc8\xc8\x03\x25\x69\x75\x26\xfc\xf5\xa9\x4c\x5c\xa8\xcc\x64\xfe\xa9\xb2\xf9\xfa\x75\x37\x8e\x42\x07\x7c\xcd\x0d\xb4\x5d\x44\x75\xc8\x81\xf7\x29\x53\x01\x10\xe7\x7c\x4c\x66\x79\xd0\xf7\x38\x9d\x6d\x02\xd3\xf9\xd2\x36\x5a\x5e\x05\x27\x8c\x2f\xd0\x5d\x15\xdc\x0c\xec\xc3\xff\x1e\x4f\xf9\x39\xa5\xc8\xab\x60\xd7\x04\x1d\xe3\x8f\xe7\x57\xc7\xff\x73\x3a\xbb\xfa\x78\x79\x75\xfc\xdf\xa7\x93\x19\x93\xdf\xdd\xa9\x02\x0c\xc2\x90\xfa\x29\x18\xc1\x20\x9d\xee\xee\xae\x76\xca\x84\x02\x76\xd2\x00\xe7\x2a\x27\x82\x7d\xf8\x8b\xdc\x89\xc4\x1d\xe1\x00\xd0\xc8\xee\x57\x62\xc7\x3d\x17\x35\x4f\xdf\xe1\x58\x61\x45\x95\xe5\x3e\xfc\x65\x38\x2a\xe0\xe4\x70\x27\x6d\xfb\x3e\xe7\xd8\x98\x3d\xc1\x5a\x52\x83\xb7\xc9\x38\xee\xfa\x86\x33\xff\x64\xc8\x65\xd9\xc5\xe1\xf4\x5f\x19\xe0\x67\xce\x00\xbd\x7f\x9b\x2b\xb3\x3b\x17\xbe\x8c\x3f\x2f\x0e\xa7\x30\x38\x27\xc0\xf0\xd4\xa7\x8d\x94\xb8\x6e\x9f\x02\x52\x24\xc3\xa7\x70\xf9\x34\x40\x22\x23\x1d\xab\xda\xfd\xbd\x71\x5d\x9b\xfd\x17\x40\x49\xcb\xb6\xc2\x6a\x9f\xe2\x78\x31\x7f\x01\x7c\xb4\x4c\x29\x6b\xac\xb9\x7e\x0f\x1c\xf7\x12\xe8\xdf\x99\x30\xb3\x13\xa7\xe4\x31\xbf\x11\x1a\x3f\xc3\xb3\xaf\x1e\xf4\xeb\xab\xe7\x78\xf5\xd5\x33\x7c\x4a\x44\x9d\xbf\x9e\xeb\xe5\x57\x30\xa8\x11\xaa\x5a\xbd\x44\x0a\x8c\x1a\x94\x57\x37\xad\x77\x4f\x5e\xc2\xb9\x89\x69\xe1\xd5\x17\xec\xb8\xfe\x13\x9c\x3b\xd5\x8d\xab\xfe\x95\x3b\x7f\xee\xdc\xb9\xbb\x0d\xb1\xe9\xe1\xc1\x6c\xf2\x0e\x06\x83\xdf\xed\x7c\xc0\x2d\xc8\x37\x78\xeb\x48\x4c\x34\xf8\xde\xbd\xe5\x58\xd9\x3c\x85\xb5\x8e\x3c\x15\x22\x4f\x00\xf8\x19\x48\xec\x38\x52\x49\x32\xa8\xd1\x71\x54\xbe\x08\x2c\x3b\xd6\x15\x56\x5c\x3d\xbc\x48\x55\xb2\x66\x1b\xaa\x7a\xcd\xf6\x9f\x80\x4c\x9e\xe2\x1c\x8a\x90\x97\x20\xd1\xe7\x4e\xcd\x53\xf0\x6f\x8f\xe3\xdb\xce\xf2\xe0\xf3\x14\x22\xf5\xfd\x97\x5d\x59\xcb\xe7\x45\x61\xde\xc9\x6b\x31\x70\x1f\xde\x86\xbb\x6f\x7e\x1f\x14\x51\xbc\x46\xf0\x4f\x8f\xde\xcd\xc3\xfd\x29\xec\xf6\xe0\x3f\xed\x3c\xbe\x4e\x61\xef\xe4\xc2\xf0\x80\x41\x85\x12\xf9\x35\x19\xbf\xf3\x4f\x1e\xab\xc4\x17\x6b\xba\x97\x26\x70\x4e\xcf\x5e\x1f\x5c\x9e\xbf\x21\x53\x6c\xf1\x19\xc3\x4e\x82\x1b\x41\x5e\x62\xb1\xd3\xca\xfa\x2f\xca\x9a\xff\x98\x18\x66\xb1\x2d\x81\x73\xf1\xce\xbd\x31\x64\x3b\xd6\xf3\x35\xe6\xaa\x50\x28\xe1\x77\x3b\x8f\x69\x3b\xfe\xe9\xc0\xa6\x17\x1b\x4c\x45\xcf\xe4\xda\x10\xea\x9b\x29\xe6\x7a\x5e\xb9\x39\x95\xfc\x01\x63\x88\x69\xb0\x4e\x2c\xf0\x07\x4c\x1f\x7a\xff\xc0\xe4\xf0\xb1\xb9\x61\xd6\x83\x33\x9b\x8b\x58\xe7\x81\x5f\xf9\x80\xd5\x30\xe3\xa5\x74\x90\x08\xe3\xcf\xa5\x0a\xa8\x95\xe7\x09\x1c\xcf\x01\x37\x26\xfe\xb5\x08\xa5\x87\x65\xa9\xf2\xb2\x45\xb0\xf2\x20\xb4\xb6\x4b\x94\x69\x32\x8a\x3e\xce\x13\xe3\xe2\x91\x5a\x8f\x87\x86\xbb\xa4\xc5\xbb\xd9\xec\x22\x49\xec\x5e\xab\x07\xcb\x6f\x3a\xb5\x15\x12\xea\x66\xae\x55\x0e\xb1\x89\x4d\x73\xab\x25\xce\xe1\x46\x09\x10\x70\x72\x3c\x6b\xff\x3f\x30\xcc\x36\x58\x8d\xb7\x46\x89\x94\xbc\xca\x10\xea\xd7\xfe\xcd\xe6\x8e\x47\xdf\xbd\x3c\x34\x99\xbb\xf7\xdf\x86\x18\xee\xd3\xb7\xe3\x75\x3a\x91\xed\x9b\x8b\x17\xfe\xdf\xc3\xbd\x7f\x23\xbc\xd4\x60\xbe\x07\x93\x94\xe1\x91\xa7\xb3\xe4\x80\xf6\x9f\x52\xac\xc3\xf4\x2d\xd4\xce\xde\x28\x89\xce\x83\x6f\xf2\x12\x84\x87\x0f\xca\x28\xdb\xbe\x3d\x99\x60\x5d\x66\x3d\x38\x41\x83\x4e\xe5\x64\x8c\x1e\x7b\x76\x6d\x10\x4e\xab\xb4\x08\x70\x6c\x64\x6d\x95\x89\xd2\xe3\x52\xab\x72\xfc\xb5\xa9\x5c\x9c\xce\x6f\x38\xf3\x21\x1b\xff\xbc\xf3\xf7\x6c\xba\x54\x45\x78\x58\xef\x4f\x1e\xdd\xf9\x23\xa3\x54\x80\x83\x26\x94\xfc\xe2\x22\x0e\x4f\xd1\x08\x13\x36\xa8\xe3\x42\xfa\x5b\x41\x9b\xe9\xba\xe7\xff\x1f\x00\x00\xff\xff\x6e\xd4\xa9\x48\x2d\x28\x00\x00")

func configDefaultConfigYamlBytes() ([]byte, error) {
	return bindataRead(
		_configDefaultConfigYaml,
		"config/default-config.yaml",
	)
}

func configDefaultConfigYaml() (*asset, error) {
	bytes, err := configDefaultConfigYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "config/default-config.yaml", size: 10285, mode: os.FileMode(420), modTime: time.Unix(1519410552, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _configGridengineTemplateTxt = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x7c\x90\xcf\x4a\xc3\x40\x18\xc4\xef\x79\x8a\x31\xb5\xc7\xdd\xe4\x05\x3c\x35\x52\xbc\x78\x10\xc1\xa3\x24\xcd\xb7\x76\x49\xb2\x1b\xf6\x8f\x8a\xcb\xf7\xee\xb2\xdb\x22\x14\x6a\x6f\xc3\xf0\x9b\xdf\x61\x36\x77\xcd\xa0\x4d\x33\xf4\xfe\x58\x6d\xee\x21\x9e\x91\x92\x7c\xed\xfd\xf4\x34\x32\x97\xc6\xe6\xe6\xcd\xba\xa9\xd3\x8e\xb9\x51\xd1\x18\x9a\x85\x0f\xa3\x8d\xa1\x00\xf4\x1f\x40\xce\x55\x29\x69\x05\x43\x90\xbb\x35\x7a\xb4\x10\xcc\x55\x4a\xab\xd3\x26\x28\xd4\x79\xbe\x12\x96\x55\x63\x3b\xd6\x27\xa8\x00\x02\x64\xc6\x92\xce\xf3\x97\x7e\xd9\x0f\x68\xe5\x35\xc3\x8c\xe3\xfb\xe7\x42\xcb\xc3\x56\xb6\x6a\x5f\x9f\xe1\xeb\x9e\x4e\xfb\xe9\xa6\x48\x79\xfd\x43\x7f\xa6\x13\x7e\xa1\xaa\x52\x92\x8f\xdf\x74\x88\xa1\x1f\x66\x62\xc6\x97\x75\x13\x39\xb8\x68\x20\xc4\xc1\x1a\xa5\x3f\xf2\x23\xbb\x92\x98\x21\x44\xc8\x7f\x76\x17\xcf\xfe\x06\x00\x00\xff\xff\x43\xce\xa0\xb4\x78\x01\x00\x00")

func configGridengineTemplateTxtBytes() ([]byte, error) {
	return bindataRead(
		_configGridengineTemplateTxt,
		"config/gridengine-template.txt",
	)
}

func configGridengineTemplateTxt() (*asset, error) {
	bytes, err := configGridengineTemplateTxtBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "config/gridengine-template.txt", size: 376, mode: os.FileMode(420), modTime: time.Unix(1514578149, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _configHtcondorTemplateTxt = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x7c\x8e\xcd\x6e\xea\x30\x10\x85\xf7\x7e\x8a\x11\xd2\x5d\x3a\x97\x17\xc8\xe6\x42\x84\xd8\x5c\x24\x1a\xf5\x67\x15\x85\x64\x12\xac\x38\x63\x18\x7b\x42\xab\xc8\xef\x5e\x05\x10\x15\x55\xe9\xee\xd8\xe7\x3b\x9f\x46\xc8\x0c\xc8\x1e\x21\x85\xa1\x24\x63\x6d\xa9\x5a\x0c\x48\x03\xa4\x90\xb3\xa0\xc2\x77\xac\x24\x94\x3b\x3b\x21\xe3\x98\x64\xb7\x77\x8c\xaa\xe4\x56\x7a\xa4\xe0\x21\x85\x93\xe3\x0e\x19\x58\x08\xb4\xae\x1c\x35\xa6\x9d\xf8\xc5\x39\xc5\x08\x5a\x87\xd2\x77\xeb\xe5\xf4\x99\x4f\xa9\x8e\x51\x59\xd7\x5e\xb4\x2f\x8e\xbb\xa5\xe1\x18\xff\x56\x8e\x6a\xc7\x1a\x07\xa4\xa0\xad\x6b\x15\x32\x3b\xfe\x4e\x35\x42\x84\x56\xfb\x50\x23\xb3\x72\x12\x0e\x12\x1e\x33\x4e\x82\xf2\x7b\x27\xb6\x2e\x02\x97\xe4\x1b\xe4\xa2\x31\x16\xa7\xbb\xdf\xb2\x27\x75\xda\x23\x15\xc1\x7d\x95\x37\xe1\xe6\x7f\x91\xbd\xae\xf3\x62\xb3\x2d\xb2\xe7\xf5\x22\x57\xe3\x68\x1a\x20\x84\x64\x71\x10\x0f\x73\xd0\x31\xaa\x71\x3c\xb0\xa1\xd0\xc0\x8c\xf1\x28\xe8\x43\x51\x4d\x65\x0a\x7f\xea\xd9\x05\x3c\x43\x1a\x90\xea\x73\xba\x2a\xb6\x65\xbf\xda\xc1\x3c\x79\x64\xe9\xb1\x77\xfc\x31\x79\x92\x79\x03\xab\x7f\xb3\xeb\xe4\x67\xdb\xd2\xf8\xee\x57\x5d\x6d\x7c\x77\x27\xbb\x2c\xee\x6c\xea\x28\x28\xa8\x3e\x03\x00\x00\xff\xff\xbb\x66\xa4\x8a\x17\x02\x00\x00")

func configHtcondorTemplateTxtBytes() ([]byte, error) {
	return bindataRead(
		_configHtcondorTemplateTxt,
		"config/htcondor-template.txt",
	)
}

func configHtcondorTemplateTxt() (*asset, error) {
	bytes, err := configHtcondorTemplateTxtBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "config/htcondor-template.txt", size: 535, mode: os.FileMode(420), modTime: time.Unix(1514578149, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _configPbsTemplateTxt = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x7c\xd0\x4d\x4b\xc4\x30\x10\xc6\xf1\x7b\x3f\xc5\xd8\x65\x8f\x49\xeb\x55\xe8\xc5\xad\x88\x17\x11\x15\x3c\x37\x9b\xc9\x1a\xda\x4e\x4a\x5e\x50\x08\xf3\xdd\xa5\x2f\x20\x0b\xd6\xdb\x30\xfc\xf9\x1d\x9e\xc3\x4d\xa5\x2c\x55\xaa\x0b\x9f\xc5\xe1\xe5\xfe\x0d\xc4\x33\xe4\x2c\xdf\xbb\xd0\x3f\x69\xe6\xed\xe7\xe6\xdf\x87\xf3\x7d\x6b\x3d\x73\x65\x12\x11\x0e\x22\x44\xed\x52\xdc\x12\xdc\x4b\xd0\xfb\x22\x67\x6b\x80\x10\xe4\x69\x4a\x01\x6a\x10\xcc\x45\xce\x93\xb7\x14\x0d\x94\x2b\x30\x00\x39\x8d\xa1\xb9\xbd\x9b\x26\x6a\x8e\xba\x5c\xeb\xa5\x14\x80\xa4\x97\x6b\x73\x5e\xbb\xf1\x51\x41\x2d\xf7\xa8\x11\xc7\xe6\x28\x6b\x73\x51\xe5\x16\xff\xed\xb4\x36\xf4\xff\x42\xc6\x0e\xf8\x2b\xad\xf9\x15\x55\xe4\x2c\x1f\xbe\xf1\x9c\x62\xa7\x06\x64\x86\x2f\xe7\x7b\xf4\xe0\x13\x81\x10\x67\x47\xc6\x5e\xe6\x69\x4e\xcb\xc5\x0c\x42\xc4\x79\xdc\xf6\x6a\xe6\x9f\x00\x00\x00\xff\xff\xd8\xa6\xd9\x67\x87\x01\x00\x00")

func configPbsTemplateTxtBytes() ([]byte, error) {
	return bindataRead(
		_configPbsTemplateTxt,
		"config/pbs-template.txt",
	)
}

func configPbsTemplateTxt() (*asset, error) {
	bytes, err := configPbsTemplateTxtBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "config/pbs-template.txt", size: 391, mode: os.FileMode(420), modTime: time.Unix(1514578149, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _configSlurmTemplateTxt = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x84\x91\xc1\x6a\xeb\x30\x10\x45\xf7\xfe\x8a\x79\x0e\x59\xca\xf6\xfb\x84\xc6\x2e\x69\xb7\x6d\xa0\x6b\x39\x1e\xb7\xaa\xe3\x91\x18\x49\xb4\x20\xe6\xdf\x8b\x9c\x40\x1c\x68\xe8\xee\x22\x1d\x1d\x34\x73\x37\xff\xea\xde\x50\xdd\x6b\xff\x51\x6c\x5e\x77\x0f\x87\xf6\x09\x94\xfa\xb4\xbd\x22\x3d\x23\xa4\x54\x1d\xb4\x9f\x9e\x07\x91\xd5\x35\x05\xed\x27\x0f\xff\x57\x47\xc8\x6c\x39\xe3\x6f\x96\xa7\xce\xb0\x48\x3d\x46\x22\x3c\x29\x1f\x06\x64\x5e\xa1\x36\x06\x17\xc3\x3d\xd6\xc6\x50\xa4\x64\x46\x20\x84\xaa\x75\xd1\x43\x03\x4a\xa4\x48\xc9\xb1\xa1\x30\x42\x79\x35\x1d\x5d\xf4\xca\x21\xab\xfc\x1f\xd8\x0e\xe5\xf9\xc5\x42\x2b\x40\x1a\x96\x74\x71\xbd\xe8\x79\xdf\x43\x53\xdd\xd7\xcd\x38\xc3\xb6\x6a\xc6\xfd\xae\xbc\xe0\xbf\x9b\x3a\xe3\xa7\x3f\x54\x61\x76\x57\xd5\x99\xbf\x71\x15\x29\x55\x8f\xdf\x78\x8c\x41\xf7\x27\x14\x81\x2f\xcb\x13\x32\x70\xa4\x3c\x97\xa5\xd1\xbc\xe7\x0d\xb5\x4b\x12\xc9\xca\xdc\x43\x77\xd3\xc8\x4f\x00\x00\x00\xff\xff\x46\x53\xad\xe9\xbd\x01\x00\x00")

func configSlurmTemplateTxtBytes() ([]byte, error) {
	return bindataRead(
		_configSlurmTemplateTxt,
		"config/slurm-template.txt",
	)
}

func configSlurmTemplateTxt() (*asset, error) {
	bytes, err := configSlurmTemplateTxtBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "config/slurm-template.txt", size: 445, mode: os.FileMode(420), modTime: time.Unix(1514578149, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"config/default-config.yaml":     configDefaultConfigYaml,
	"config/gridengine-template.txt": configGridengineTemplateTxt,
	"config/htcondor-template.txt":   configHtcondorTemplateTxt,
	"config/pbs-template.txt":        configPbsTemplateTxt,
	"config/slurm-template.txt":      configSlurmTemplateTxt,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"config": {nil, map[string]*bintree{
		"default-config.yaml":     {configDefaultConfigYaml, map[string]*bintree{}},
		"gridengine-template.txt": {configGridengineTemplateTxt, map[string]*bintree{}},
		"htcondor-template.txt":   {configHtcondorTemplateTxt, map[string]*bintree{}},
		"pbs-template.txt":        {configPbsTemplateTxt, map[string]*bintree{}},
		"slurm-template.txt":      {configSlurmTemplateTxt, map[string]*bintree{}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}