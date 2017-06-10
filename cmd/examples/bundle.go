// Code generated by go-bindata.
// sources:
// examples/config.yml
// examples/file-contents.json
// examples/google-storage.json
// examples/hello-world.json
// examples/log-streaming.json
// examples/md5sum.json
// examples/port-request.json
// examples/resource-request.json
// DO NOT EDIT!

package examples

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

var _examplesConfigYml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xac\x58\xdf\x73\xdb\xb8\x11\x7e\xe7\x5f\xb1\x13\x3d\xf4\x6e\x46\x62\xe8\xb9\xb9\x4e\x4f\x6f\x8e\x9c\xf8\x3c\xe7\xb4\xaa\xed\x8c\x67\xfa\x06\x02\x2b\x12\x35\x08\xb0\xc0\x52\x0a\xf3\xd7\x77\x16\x00\xf5\x23\x91\x13\xb7\xbe\x37\x11\x58\x7c\xbb\x00\x76\xbf\x6f\xa1\x19\xac\x05\xb5\x40\x0e\xa8\x45\x50\x82\x44\x2d\x02\xc2\x46\x1b\x2c\x8b\xab\x77\x3c\xb9\x84\xf2\xed\x66\xb0\x16\xcd\x62\xe7\xfc\xd3\x42\x69\x9f\xbf\x4b\x55\x17\x33\x58\x3b\x4f\x30\x04\x54\xb0\x71\x1e\x7e\x7f\x78\x58\x83\x74\x5d\x37\x58\x2d\x05\x69\x67\x41\x58\x15\xd1\x77\x58\x83\x12\xa1\xad\x9d\xf0\xaa\x2c\xd8\x92\xd7\x2e\xe1\x6f\x55\x55\x7d\x03\x74\xb7\x5e\x9d\xe2\x94\xc5\xdd\x7a\x95\x16\xfc\x56\xfd\xc6\x0b\x1e\x9d\x7f\xd2\xb6\x89\xd1\x06\x90\x1e\x05\xa1\x02\x35\x78\x1e\xec\xbd\x93\x18\x02\xff\xdc\x69\x63\xa0\x46\xd8\x79\x4d\x84\x16\xb4\x05\x6a\x75\x00\xa5\x3d\x4a\x72\x7e\x2c\x0b\x86\xba\xd2\xfe\xcc\x5e\x8b\x19\xfc\xee\x02\x59\xd1\x21\xb8\x4d\xdc\xc8\x87\x68\x01\x01\xfd\x16\x7d\x09\x1f\x85\xb6\x66\x9c\x27\x4c\x1d\xd2\x16\xea\x11\xc4\x40\x6e\x11\xa4\x30\xe8\x43\x31\xe3\x33\x96\xce\x6e\x74\x33\x78\x04\x46\x47\x1f\xca\x82\xb1\xff\x2e\x3a\x5c\x82\x71\x52\x98\xd6\x05\x2a\x66\x70\xeb\x9a\x86\x23\x37\xb8\x45\x13\x96\xa0\xb0\x1e\x9a\x39\x68\xbb\x71\x73\x40\xef\x9d\x2f\x6e\x5d\x73\xcb\xb3\x79\x92\x8f\xc3\x6b\x42\x30\xae\x09\xe9\x3e\x75\x80\x5e\x50\x5b\xc2\xcd\x06\xb0\xeb\x69\x9c\xa7\x49\xe1\x0f\x47\x41\x0e\x02\x29\xf4\xbe\x64\xc0\x74\xdd\x6f\xde\x70\x04\xba\xd3\x14\x77\x1b\xf4\x97\xb4\x73\x11\x9e\x00\x3f\xa3\x1c\xc8\xf9\x84\xf4\x53\x20\xe5\x06\x7a\x8b\xde\xff\xcc\xd1\x41\x3d\x12\x86\xb2\xf8\x28\x3e\xbf\xcf\x86\xb7\xae\xb9\xd7\x5f\x70\x09\x17\x55\x55\x55\x30\x83\x8b\x0a\xfe\x78\x17\x0f\x75\x07\x6e\x93\x63\xf0\x83\x05\x01\x41\xb6\xa8\x06\x83\x1e\x34\xa1\xcf\x97\x3e\x83\x1b\x0b\x56\x58\x17\x50\x3a\xab\x42\x59\xdc\x67\xb3\x3b\x41\x13\x6e\x35\x81\x43\xb2\xca\xf8\x9d\xb0\x63\x8c\x3b\x9e\xc8\x04\xcf\x81\x3a\x8b\xc7\x4e\x26\xc8\x55\x3b\xd8\x27\xc6\xcc\x00\xc6\xd9\x86\x97\xee\x84\x26\xa8\x91\x76\x88\x16\x86\x5e\x09\xc2\x00\x35\x6e\x9c\x47\xe8\x44\xca\x42\x91\x2f\x15\x14\x0a\x75\x2e\xee\xc7\x38\xbd\xd6\xb6\x79\xd0\x1d\xba\x81\x96\xf0\xd7\xea\x34\xfa\x4e\xdb\x81\xf0\x9c\x73\x2e\x89\xbd\x87\x78\x6b\xc2\xd3\xfc\xeb\x18\x62\x99\xbd\x24\x8a\x1b\xab\x69\x1f\xc5\x2f\xd5\x49\x18\xbf\xe6\x30\x42\x5c\x2d\xcd\xa0\x10\x04\xbc\x59\x09\xd9\xe2\x62\xe5\x2c\x79\x67\x96\x60\xdd\x22\x90\xf3\xf8\x26\x95\x7c\x8b\x42\xf1\xc5\x59\xb8\x46\x7a\x7b\xab\x03\x81\xc7\xd0\x3b\x1b\x30\x27\x7f\xef\x71\x8b\x96\x40\x0a\xd9\x72\xac\xf5\x08\xda\x12\xfa\x0e\x95\x16\x7e\x8c\xe5\xa4\x25\xa7\xcf\x95\x0e\xa2\x36\xc8\xb8\xd1\xe9\x12\xc8\x0f\x58\x14\x33\xf8\xa0\x0d\x02\xbb\x15\x0d\x42\x18\x03\x61\xc7\x09\x91\x06\x96\x05\x00\x97\x8e\x14\x26\x52\x42\x36\x28\x0b\x48\x83\x3c\xcf\x16\x8f\xad\x26\x34\x1c\xa1\xdb\xa4\xaa\x3b\x30\x41\xac\x98\x00\xbb\x56\xcb\x76\xaa\x73\x1d\x40\x18\xe3\x76\xa8\x78\x1b\x42\x32\xad\x94\x11\xeb\x32\x0d\x5f\x69\x1f\x12\x38\xc0\xe2\x0c\x5b\xe6\x80\x8b\x18\xdf\xfd\x2f\x29\x4e\x36\x7d\x6f\x55\xef\xb4\xa5\x69\x04\xe0\x0f\x1c\x0f\x1f\xf7\x28\x3d\xd2\x32\xad\xbb\x76\xae\x31\x08\x2b\xe3\x06\x05\x79\xc7\x69\xe2\xfe\x00\x78\xa0\x73\x21\xa5\x1b\xf8\xb4\x3d\x2a\xb4\xa4\x85\x09\x99\xd5\x27\xf4\x19\xfc\xa3\xe7\xec\x17\x26\x32\x44\xef\x42\xd0\xb5\xc1\xcc\x63\x13\x67\x32\x89\x75\x82\xb4\x14\xc6\x8c\xa0\x74\x90\x6e\x8b\x1e\xd5\x01\xe7\x32\x79\xe2\xab\x59\x1e\x81\x5f\x9e\x5d\x78\x1a\x8f\x77\x5d\x4c\x59\xb4\x5b\xed\x9d\xed\xd0\xd2\x01\xf7\x83\x77\xdd\x7b\xbb\x3d\x5c\xfe\x43\x8b\x70\x4c\xc3\x42\x92\xde\xe2\x11\x6b\xd4\x42\x3e\x61\x2c\xfe\xcb\xad\xd0\x86\x73\x68\x1a\x0b\x99\x5f\xe7\xcc\xc1\xca\xf9\x39\x34\x12\xe7\xe0\x7a\xb4\x81\x84\x7c\xda\x33\x80\xcf\x86\xec\xf0\xfe\x6b\xe4\x4c\xe0\xc5\xbb\x09\x94\xaf\xe6\x7a\xf5\x7e\xca\xac\x97\x1c\x3e\xdb\xfd\x7f\xc7\x7d\x7a\xd4\x4c\xd3\x09\xed\x24\x33\x7a\xef\xfe\x8d\x92\xe0\xe6\xea\xb5\xce\xd6\x09\xe9\x39\x47\x5f\x9c\x7d\xf5\x7e\xfe\xe5\xec\x61\x23\x8f\xa8\x9b\x96\xf6\x85\x34\x83\xb5\xc7\x0d\xfa\x49\x2a\x81\x5a\x41\x89\xf6\x60\xe8\xe1\x3f\x83\x96\x4f\x66\x2c\xf7\xd6\x8f\xc7\x66\x2c\x71\xc2\x78\x14\x6a\x04\x67\x8d\xb6\x08\xad\xd8\x32\xf3\x07\x12\x36\xc3\x0c\x3d\x90\xee\x70\x82\x48\xee\xfe\xc9\xb8\xf7\x69\x7a\x09\x17\x65\x95\xb7\x78\x4c\xc8\xcc\x60\xc8\xf7\x0e\x97\xeb\x1b\xa6\xb9\xc1\x50\x80\x9f\xba\x48\x6c\x08\x4c\x2d\x73\x20\xec\x7a\xc3\x3a\x31\x07\x24\xf9\x73\x86\xc9\x7c\xed\x71\xe3\x31\x30\x0d\x26\xef\x91\xe6\x1e\x1e\x6e\x9f\x55\x84\x4c\xdb\xa9\xc0\x0e\x7a\xbe\x57\x84\x9b\xab\x67\x8e\xbc\x41\xcb\x2a\x97\x4e\xfc\xe6\x2a\x9f\xf6\x2c\xb6\x57\x42\x29\x8f\x21\x9c\x6d\x6b\x0a\xa6\x1f\xfe\x71\x99\x8c\x8e\xfa\x93\x65\x6c\xbf\x18\xe4\xc3\xeb\x1b\x2f\x80\xef\xb4\x5e\x79\xb3\x47\xaa\xd6\x0a\xd6\x5d\x86\x52\x06\xa3\x28\xf2\xa5\xb0\x24\xb6\xc2\x46\x43\x4a\xaa\x36\x07\x4d\xc9\x7f\x68\x07\x02\xe5\x76\x36\x51\xcb\xe2\x02\x3a\x14\x96\x13\x05\x3d\x32\xb9\x5b\x37\x2d\x2a\xa1\x9a\x26\xd3\x00\xe8\x2e\xca\x13\xa1\x19\x41\x6c\x28\x3a\xe2\xf6\xd8\x07\x8a\x8d\x05\x63\xee\x75\x74\x71\x11\x3d\x7c\x14\x9f\x75\x37\x74\xa9\x61\x32\xae\x39\x69\x93\x62\x3b\x75\xdc\x2b\xb1\x38\x35\x0f\x42\x9b\x73\x6d\xd2\xfe\x94\xbf\x15\x3d\x80\x59\x9c\x7d\xd8\x77\x9d\x43\xbc\xf0\x80\xf4\x55\x0a\xd4\x63\xea\xe3\x26\x36\x9b\x43\x3d\x10\x8c\x6e\x80\x8e\x2b\x0e\x2c\xb2\xb2\xb5\x3a\x44\x3c\xbd\xe1\xa9\xbf\x78\x4c\x45\x72\xd2\xd9\x74\xc2\x46\x27\xec\xfd\x48\x74\x9f\x95\x5d\x96\xa4\x23\xe9\xfd\xb3\xe5\xf7\x19\x09\xfe\x91\x0c\xe7\x40\x92\x10\xa7\x98\x4e\x94\x38\x0f\x65\x29\xce\x5f\x47\x5a\x0c\x67\xd8\xf0\xfe\x04\x1b\x66\x49\x94\xf7\x4c\xf6\x22\x5d\xf8\xae\x28\x1f\x2f\xf9\x11\xb9\x1e\xa1\x3d\x27\xb1\x93\x48\x9f\x6a\x37\xbc\x5e\xbb\x33\xc2\xa9\x78\xa7\x04\xf9\xe1\x93\x26\x96\xc2\xc9\xa3\x26\xe6\xcb\xeb\x9e\x35\x11\x74\xff\xb0\x89\x27\xbc\x45\xef\x35\x37\xb5\xfb\x2e\xc1\x63\x70\x83\x97\xa9\x1e\xef\xa6\x8f\x29\xb7\x57\xeb\x4f\xe1\x60\x3c\x89\xde\xaa\x1f\xc2\x12\xaa\x22\x7f\xde\x5d\x7e\x3c\xd8\xc4\x02\xbf\x7e\x37\x99\xde\x89\xee\xba\x5e\x42\x55\xee\xad\xaf\x74\x78\x82\xd0\x0b\x89\xcf\x2c\x62\x03\x5e\x73\x51\x55\x65\x26\xdb\x48\x76\xbb\x45\x3c\x40\xa0\xc1\x66\xf5\x48\x5c\x3e\x11\x16\x53\x62\x7a\xa0\xbc\x0d\xa3\x95\xc0\x57\x98\xac\xbe\x7e\x08\x00\x7c\x8a\x76\x7b\xfa\x3a\xff\x92\xfa\xbe\xeb\x93\x57\x5c\x74\xb8\xd3\x9c\xeb\xdf\xbc\x92\xbf\x1b\x42\x7a\xc9\xfd\x7a\xfa\x08\xc9\x56\xff\x53\x00\xdc\xa3\xed\x69\x77\x7a\xa7\xe5\x3f\x37\x5e\x14\xcf\xad\x6b\x5e\x10\xd2\x7f\x03\x00\x00\xff\xff\x5f\x4c\x11\x01\x36\x11\x00\x00")

func examplesConfigYmlBytes() ([]byte, error) {
	return bindataRead(
		_examplesConfigYml,
		"examples/config.yml",
	)
}

func examplesConfigYml() (*asset, error) {
	bytes, err := examplesConfigYmlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "examples/config.yml", size: 4406, mode: os.FileMode(420), modTime: time.Unix(1496339225, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _examplesFileContentsJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x8c\x91\x41\x6f\xe2\x30\x10\x85\xef\xfc\x8a\xa7\x5c\xb8\x20\x72\xda\x0b\xe7\xdd\x55\x91\x7a\xeb\xa1\x87\x16\x21\x37\x9e\x80\x25\xdb\x13\xd9\x63\xb5\x08\xf1\xdf\x2b\x3b\x09\x55\x53\x0a\xe4\xe4\x68\xe6\xbd\x99\xf9\xde\x71\x06\x00\x95\x57\x8e\xaa\x15\xaa\xb5\xef\x92\xa0\x35\x96\xd0\xb0\x17\xf2\x12\xa1\xbc\x06\x27\x19\x0b\xd5\xa2\x97\x68\x8a\x4d\x30\x9d\x18\xf6\x59\xf9\x97\x1c\xfb\x28\x41\x09\x45\xa4\x68\xfc\x0e\xb2\x27\xcc\x47\x9b\x39\x5a\x43\x56\xa3\xe5\x00\x93\xa7\x44\x08\xa3\x09\xa4\x84\xa0\xfa\x91\xec\x8b\x66\xcf\x51\x10\x0f\x51\xc8\x8d\xc3\x7a\x45\xb5\xc2\x4b\xf9\xcf\xdf\xf1\xfc\xfa\x76\x41\xa3\xa4\xf7\x1f\xa4\xe7\x86\xc9\xbe\xfd\xa5\xc2\x70\xfa\x4f\x4c\x6e\x89\x5a\x5c\x57\x1b\x8f\x77\x63\x2d\xde\x68\x58\x4d\x5f\x58\x6a\x39\xb5\x96\x43\x57\x66\xff\x5f\x3f\xfe\x9b\xd6\x3a\x25\xfb\x5c\x1b\xec\xa7\xe5\x11\x4f\x6e\x79\x20\x6b\x19\xcf\x1c\xac\x7e\xf5\xd5\xb9\xef\x54\x5e\x9b\x81\x44\x9f\xc4\xdd\x28\xa2\x68\xbe\xc9\xe2\xa9\x34\x81\x5b\x14\x7a\x11\x8d\xea\x24\x05\x2a\x11\x95\xc5\x85\xa2\x6c\x4b\xcf\x6d\x1a\x29\xd8\x6c\x9a\x13\x5d\xd5\x75\x91\x37\xaa\xa8\x2f\xa4\x72\x37\xba\x7c\xc5\x2f\x48\xe8\x83\x9a\x24\x1c\xae\x42\x31\x4e\xed\x68\x3b\xa2\x51\xb6\x33\x9e\x7e\x84\xe1\x74\xb6\xc8\xe0\xaa\xc5\x57\x64\x9b\x49\xdb\xc0\xf4\xda\x66\xb3\xd3\xec\x33\x00\x00\xff\xff\x3f\xad\x1a\x6c\x5a\x03\x00\x00")

func examplesFileContentsJsonBytes() ([]byte, error) {
	return bindataRead(
		_examplesFileContentsJson,
		"examples/file-contents.json",
	)
}

func examplesFileContentsJson() (*asset, error) {
	bytes, err := examplesFileContentsJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "examples/file-contents.json", size: 858, mode: os.FileMode(420), modTime: time.Unix(1495154596, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _examplesGoogleStorageJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x6c\x90\x4f\x4b\x03\x31\x10\xc5\xef\xfd\x14\x8f\x9c\x14\x6a\x77\x3d\x54\xb0\x57\x05\x2f\x82\x62\xed\x49\x44\xa6\x49\x76\x0d\xe6\x1f\x9b\x09\x96\x96\x7e\x77\xc9\x76\x17\xb6\xd2\x53\x86\xcc\xfc\xde\xbc\x79\x87\x19\x00\x08\x4f\x4e\x8b\x15\xc4\x53\x08\xad\xd5\x78\xb0\x21\x2b\xac\x39\x74\xd4\x6a\xe8\x1d\xb9\x68\xb5\x98\x9f\x66\x95\x4e\xb2\x33\x91\x4d\xf0\x05\x79\xa7\xf4\x03\xe3\x63\xe6\x04\xf2\x0a\x21\x73\x5f\x4b\xf2\xd8\x6a\x5c\x54\xdc\xbc\x3d\xa7\xc5\xa8\xa7\x77\x5a\x66\x0e\x5d\x12\x2b\x7c\xf4\x5f\xc0\x61\x78\x01\x61\x1c\xb5\xfa\x6b\x34\x98\xb7\xd9\x73\x1e\xd0\xbe\x2f\x9d\x2a\xa0\x70\x6a\x99\xb2\x13\x73\x88\x8a\x5d\xac\x1a\x63\xf5\x62\x6f\xa2\xf8\x9c\xcc\x26\x56\x21\x73\xd1\x19\xaa\xa1\x75\xec\xdf\x61\x52\x9c\x8e\xb9\xec\x66\xf4\xb1\x37\xb1\x6c\x98\x1a\xf9\x97\xcb\x63\xf8\xf5\x36\x90\x02\x21\xe6\xad\x35\x12\x05\x40\xd3\x05\x37\x86\x32\xc6\x71\xb5\x59\xe3\x95\x58\x7b\xc6\x4b\xd3\x18\xa9\xa1\x88\xe9\x7a\x2a\x9e\x3b\x5b\x44\xdb\xb4\xaa\xaa\x9c\x22\x87\x9b\x48\xa6\xab\x28\x46\x6b\x24\x95\x9d\xa9\xaa\x97\xf7\x75\x5d\xdf\xde\xf5\x67\x4f\xd8\x48\xfc\x5d\xe0\xf3\x5c\xce\x4f\x9f\x1d\x67\x7f\x01\x00\x00\xff\xff\xb5\x5b\x74\xc7\x0b\x02\x00\x00")

func examplesGoogleStorageJsonBytes() ([]byte, error) {
	return bindataRead(
		_examplesGoogleStorageJson,
		"examples/google-storage.json",
	)
}

func examplesGoogleStorageJson() (*asset, error) {
	bytes, err := examplesGoogleStorageJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "examples/google-storage.json", size: 523, mode: os.FileMode(420), modTime: time.Unix(1494624283, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _examplesHelloWorldJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x5c\x8e\x41\x0a\xc2\x40\x0c\x45\xf7\x3d\x45\xc8\xba\x78\x80\xae\x5d\x78\x07\x29\x12\x67\x82\x1d\x9c\x99\x94\x49\x8a\x42\xe9\xdd\x65\x6a\x5b\xc4\xac\x3e\xf9\x2f\x3f\x7f\x6e\x00\x00\x30\x53\x62\xec\x00\x2f\x1c\xa3\xc0\x4b\x4a\xf4\xd8\x7e\x2d\xcf\xea\x4a\x18\x2d\x48\xae\xc4\x99\x93\x64\xb5\x42\xc6\x0a\x36\x30\x24\x51\x83\x3b\x69\x70\xc0\x6e\x10\x30\xd2\xe7\x69\x3f\xe6\x37\xbb\xc9\xa4\x28\x76\x70\x5d\x57\x75\xe6\x43\xad\x50\x48\xf4\xe0\xdb\x5e\x81\xe2\x18\x32\x6f\x01\x07\xe3\x92\xaf\x11\x58\x5f\x60\x0b\x38\xfc\x14\xed\xff\x58\x35\x2f\x93\xd5\xac\x4d\x1d\xf6\xb2\xaa\xbe\x59\x9a\x4f\x00\x00\x00\xff\xff\x40\x5f\xc3\x4d\xf8\x00\x00\x00")

func examplesHelloWorldJsonBytes() ([]byte, error) {
	return bindataRead(
		_examplesHelloWorldJson,
		"examples/hello-world.json",
	)
}

func examplesHelloWorldJson() (*asset, error) {
	bytes, err := examplesHelloWorldJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "examples/hello-world.json", size: 248, mode: os.FileMode(420), modTime: time.Unix(1494624283, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _examplesLogStreamingJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x64\x90\xcb\x6a\x23\x31\x10\x45\xf7\xfe\x8a\x8b\xd6\x1e\x37\xb3\xb5\xb7\xc3\xac\x66\x31\xfb\x60\x82\x2c\x5d\xba\x05\x7a\x34\x55\xd5\x49\x8c\xf1\xbf\x07\xb5\x1f\x84\x44\x0b\x51\x8f\xc3\x51\xa9\x2e\x1b\x00\x70\xd5\x17\xba\x3d\xdc\xbf\x36\x42\x4d\xe8\x4b\xaa\xa3\xdb\xde\x9a\x91\x1a\x24\xcd\x96\x5a\xed\xcc\x1f\x96\x56\xd5\xc4\x1b\x15\x36\x79\x83\x5a\x6c\x8b\x0d\x14\x41\x6e\xa3\xc2\x0b\xef\x16\x46\x9c\xce\xb0\x06\x9b\x88\xbf\x4b\xad\xcc\x88\x5e\xa7\x53\xf3\x12\x75\x87\xff\x92\xaa\xe9\xda\x8d\xde\x38\x58\x2a\x04\xdf\x28\x67\x28\x43\xab\x71\x8b\xa5\x5a\xca\x08\xbe\x06\x66\xc6\xdd\x63\x28\x7e\x30\x2c\xd6\x44\xdd\x1e\x2f\x6b\xa9\x9f\xcb\x33\x5a\xa1\x54\xfc\xc8\xd7\xc7\xe7\x7c\x9e\x53\xe5\x5d\xf0\x64\x42\x89\x5d\xe1\x74\x72\x5b\xb8\x5f\xa1\xdf\xef\x53\xca\x84\xc9\xc2\x03\x62\x5b\x47\x3b\x40\x33\x39\xe3\x77\xaf\x54\xba\xe3\x37\xcd\x6d\x07\xfd\x99\xc1\xca\x3c\xdc\xd3\x9f\x10\x45\xbe\x42\x3d\x7d\x32\xd7\x35\x3a\x6e\xae\x9b\xcf\x00\x00\x00\xff\xff\x31\x2b\x81\x7b\x97\x01\x00\x00")

func examplesLogStreamingJsonBytes() ([]byte, error) {
	return bindataRead(
		_examplesLogStreamingJson,
		"examples/log-streaming.json",
	)
}

func examplesLogStreamingJson() (*asset, error) {
	bytes, err := examplesLogStreamingJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "examples/log-streaming.json", size: 407, mode: os.FileMode(420), modTime: time.Unix(1494624283, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _examplesMd5sumJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xa4\x91\xcd\x6a\x2b\x31\x0c\x85\xf7\x79\x0a\xe1\x75\xc8\xac\xee\x26\xeb\xdb\x42\xa0\xbb\x2e\x4b\x08\x66\xac\x64\x0c\x23\xdb\x8c\x64\x48\x08\x79\xf7\xe2\x9f\x99\x52\xb7\x4d\x69\xeb\x95\x8c\x8e\x8e\x8f\x3e\x5f\x57\x00\x00\xca\x69\x42\xb5\x05\xb5\x73\x21\x0a\x68\x67\xc0\x47\x49\xe5\xd1\x8e\xa8\xd6\x45\x64\x90\xfb\xc9\x06\xb1\xde\x25\xed\x7f\x24\xef\x58\x26\x2d\xc8\x60\x3f\x1b\x64\x88\x6c\xdd\x09\x34\xb0\xa5\x30\x22\x90\xf9\xc7\x91\xa0\xf7\x44\xda\x99\xcd\x6c\x9c\x87\x59\x6d\xe1\x25\xdf\xd3\xb9\x2e\xd5\xbb\x7c\xd5\x20\x0f\xd4\xe9\x45\xd3\xc4\x2b\xab\x88\xaf\x8f\x6e\xa0\x13\x0a\x5d\xb9\x1c\x4a\x5c\x8a\x2c\x80\x67\xcb\x02\xde\x81\x0c\x08\x83\x67\x01\xbe\xb0\x20\x6d\x5a\xff\x38\x8d\xc9\x37\x2d\xb6\xed\xba\x0f\x6e\xad\x5c\x2e\x21\x47\x7e\xdc\x3d\x3d\xb4\xbd\xa0\x65\x48\xbd\x6c\x62\x9d\x5a\xba\xb7\x5c\xed\x2b\x97\x82\xf2\x27\x60\x58\x8c\xff\x96\xcc\x73\x16\x81\x3f\xce\xff\x61\x19\x7a\x1d\x24\x4e\xc8\x09\x58\x4e\x25\xc8\x72\xc8\xb2\xbf\x90\x29\x1b\xfc\x1a\x4d\xda\xe5\x0b\x36\x78\xc6\x3e\x8a\x9f\xee\xd2\xb1\xa4\x4f\x78\x98\x19\xe9\x31\x58\x87\xed\x83\x3d\x99\x64\x51\x09\xaa\xf5\xdb\xaf\xec\x1b\x65\x85\x7b\x2f\xdc\xea\xb6\x7a\x0d\x00\x00\xff\xff\xba\xce\x0d\x5b\x52\x03\x00\x00")

func examplesMd5sumJsonBytes() ([]byte, error) {
	return bindataRead(
		_examplesMd5sumJson,
		"examples/md5sum.json",
	)
}

func examplesMd5sumJson() (*asset, error) {
	bytes, err := examplesMd5sumJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "examples/md5sum.json", size: 850, mode: os.FileMode(420), modTime: time.Unix(1494624283, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _examplesPortRequestJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x5c\x8e\xb1\x6a\xc4\x30\x10\x44\x7b\x7f\xc5\xb0\xb5\x09\x97\xce\xb8\xce\x07\xa4\x0f\x26\x08\xdf\xe2\x13\x9c\x76\x15\xed\x1e\x04\x0e\xff\x7b\x90\x12\x0b\x1c\x55\x83\xe6\xe9\x8d\x9e\x03\x00\x90\x84\xc4\x34\x83\xde\xb5\x38\x0a\x7f\x3d\xd8\x9c\xc6\xdf\xee\xca\xb6\x96\x98\x3d\xaa\x54\xe4\x8d\x93\x8a\x79\x09\xce\x86\x60\x16\x37\x89\xb2\x21\x6b\x71\x83\x2b\xfc\xc6\xe0\x6f\x5e\x1f\xae\x05\xab\x8a\x87\x28\x5c\x5e\x0e\xdb\x51\x19\xcd\xf8\x68\x57\xf5\x3c\x7b\x6a\x50\x4c\x61\xe3\xcf\xe3\x53\xe1\x9e\xa3\xf0\x9f\xa0\x33\x6d\xb0\x4a\xce\x6f\x01\xba\xa9\x39\xcd\xb8\x8c\xa0\xbe\x4f\x33\xa6\x69\x9a\x4e\xe8\xbe\xfc\x53\xae\xe9\x5a\x85\x64\x77\xe6\x4c\x23\xe8\xf5\x42\x4b\x47\xf6\x96\x96\x61\x1f\x7e\x02\x00\x00\xff\xff\x76\x8e\xbb\x27\x35\x01\x00\x00")

func examplesPortRequestJsonBytes() ([]byte, error) {
	return bindataRead(
		_examplesPortRequestJson,
		"examples/port-request.json",
	)
}

func examplesPortRequestJson() (*asset, error) {
	bytes, err := examplesPortRequestJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "examples/port-request.json", size: 309, mode: os.FileMode(420), modTime: time.Unix(1494624283, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _examplesResourceRequestJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x54\x90\xcd\x6a\xc3\x30\x10\x84\xef\x7e\x8a\x61\xcf\xa2\xd8\x90\x5c\x7c\xeb\x0f\xf4\x54\x28\x81\x9e\x4a\x30\x5b\x79\x31\x22\xb1\xa4\x6a\x65\x28\x2d\x7e\xf7\x62\xc5\x4e\xda\x9b\xf4\x69\xf6\x1b\xb1\x3f\x15\x00\x90\xe7\x51\xa8\x05\x1d\x44\xc3\x94\xac\x20\xc9\xe7\x24\x9a\xc9\x5c\xde\x7b\x51\x9b\x5c\xcc\x2e\xf8\x25\xf6\x24\x63\xf0\x9a\x13\x67\x51\x30\x32\xeb\x69\x9b\x70\x7e\xc0\x0e\x8f\xaf\x6f\xb0\x21\x89\x1a\x34\x7b\x3c\x3f\xe0\x70\xff\x62\xc0\xbe\x47\x53\xd7\xcb\xbd\x77\x7a\x82\x46\xb6\x72\xb7\x75\xa4\xb5\x5b\xa9\xc5\xe5\x5b\x05\xdb\x38\x75\x45\x45\x2d\x76\xe6\xc6\x13\x8f\xdd\xf0\x41\x2d\x9a\xfd\x1f\xaa\xee\x5b\x56\x5c\xd7\x05\xcf\xab\x5e\xbe\xc4\x4e\x39\xa4\xc5\xf3\x7e\x1d\xb8\x15\x95\x90\x1b\x79\x90\x6e\xdb\x06\x9f\xa3\xf3\x42\xe6\x7f\xc6\x8e\xfd\xa2\x20\x3d\x8b\x44\x32\xa0\x86\x8e\xd7\xc4\x5c\x4e\xc7\x6a\xae\x7e\x03\x00\x00\xff\xff\x3a\xf5\x0f\x91\x5a\x01\x00\x00")

func examplesResourceRequestJsonBytes() ([]byte, error) {
	return bindataRead(
		_examplesResourceRequestJson,
		"examples/resource-request.json",
	)
}

func examplesResourceRequestJson() (*asset, error) {
	bytes, err := examplesResourceRequestJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "examples/resource-request.json", size: 346, mode: os.FileMode(420), modTime: time.Unix(1494624283, 0)}
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
	"examples/config.yml":            examplesConfigYml,
	"examples/file-contents.json":    examplesFileContentsJson,
	"examples/google-storage.json":   examplesGoogleStorageJson,
	"examples/hello-world.json":      examplesHelloWorldJson,
	"examples/log-streaming.json":    examplesLogStreamingJson,
	"examples/md5sum.json":           examplesMd5sumJson,
	"examples/port-request.json":     examplesPortRequestJson,
	"examples/resource-request.json": examplesResourceRequestJson,
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
	"examples": {nil, map[string]*bintree{
		"config.yml":            {examplesConfigYml, map[string]*bintree{}},
		"file-contents.json":    {examplesFileContentsJson, map[string]*bintree{}},
		"google-storage.json":   {examplesGoogleStorageJson, map[string]*bintree{}},
		"hello-world.json":      {examplesHelloWorldJson, map[string]*bintree{}},
		"log-streaming.json":    {examplesLogStreamingJson, map[string]*bintree{}},
		"md5sum.json":           {examplesMd5sumJson, map[string]*bintree{}},
		"port-request.json":     {examplesPortRequestJson, map[string]*bintree{}},
		"resource-request.json": {examplesResourceRequestJson, map[string]*bintree{}},
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
