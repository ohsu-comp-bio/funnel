package storage

import (
	"context"
	"fmt"
	"io"
	urllib "net/url"
	"os"
	pathlib "path"
	"strings"

	"github.com/jlaffaye/ftp"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/util/fsutil"
)

// FTP provides read access to public URLs.
type FTP struct {
	conf config.FTPStorage
}

// NewFTP creates a new FTP instance.
func NewFTP(conf config.FTPStorage) (*FTP, error) {
	return &FTP{}, nil
}

func (b *FTP) connect(url string) (*ftp.ServerConn, error) {
	u, err := urllib.Parse(url)
	if err != nil {
		return nil, fmt.Errorf("ftpStorage: parsing URL: %s", err)
	}

	host := u.Host
	if u.Port() == "" {
		if u.Scheme == "sftp" {
			host += ":22"
		} else {
			host += ":21"
		}
	}

	// TODO we should probably optimize client connections.
	//
	//      most of the FTP code in this file is creating/using clients
	//      very inefficiently. For example, a directory download or list,
	//      is probably going to recreate the same client many times.
	client, err := ftp.Dial(host)
	if err != nil {
		return nil, fmt.Errorf("ftpStorage: connecting to server: %v", err)
	}

	// TODO should be pulling auth details from both the URL and the config.
	//      want to figure out how to manage auth creds for multiple FTP sites.
	err = client.Login("anonymous", "anonymous")
	if err != nil {
		return nil, fmt.Errorf("ftpStorage: logging in: %v", err)
	}
	return client, nil
}

// Stat returns information about the object at the given storage URL.
func (b *FTP) Stat(ctx context.Context, url string) (*Object, error) {
	u, err := urllib.Parse(url)
	if err != nil {
		return nil, fmt.Errorf("ftpStorage: parsing URL: %s", err)
	}

	client, err := b.connect(url)
	if err != nil {
		return nil, err
	}

	resp, err := client.List(u.Path)
	if err != nil {
		return nil, fmt.Errorf("ftpStorage: listing path: %v", err)
	}

	if len(resp) != 1 {
		return nil, fmt.Errorf("ftpStorage: object not found: %s", url)
	}

	r := resp[0]

	// TODO there is a "link" file type. can we support that?
	if r.Type != ftp.EntryTypeFile {
		return nil, fmt.Errorf("ftpStorage: stat on non-regular file type: %s", url)
	}

	return &Object{
		URL:          url,
		Name:         strings.TrimPrefix(u.Path, "/"),
		LastModified: r.Time,
		Size:         int64(r.Size),
	}, nil
}

// Get copies a file from a given URL to the host path.
func (b *FTP) Get(ctx context.Context, url, path string) (*Object, error) {
	obj, err := b.Stat(ctx, url)
	if err != nil {
		return nil, err
	}

	client, err := b.connect(url)
	if err != nil {
		return nil, err
	}

	fmt.Println("GET", obj.Name)
	src, err := client.Retr(obj.Name)
	if err != nil {
		return nil, fmt.Errorf("ftpStorage: executing RETR request: %s", err)
	}
	defer src.Close()

	dest, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("ftpStorage: creating host file: %s", err)
	}

	_, copyErr := io.Copy(dest, fsutil.Reader(ctx, src))
	closeErr := dest.Close()

	if copyErr != nil {
		return nil, fmt.Errorf("ftpStorage: copying file: %s", copyErr)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("ftpStorage: closing file: %s", closeErr)
	}

	return obj, err
}

// Put is not supported by FTP storage.
func (b *FTP) Put(ctx context.Context, url string, hostPath string) (*Object, error) {

	u, err := urllib.Parse(url)
	if err != nil {
		return nil, fmt.Errorf("ftpStorage: parsing URL: %s", err)
	}

	reader, err := os.Open(hostPath)
	if err != nil {
		return nil, fmt.Errorf("ftpStorage: opening host file for %q: %v", url, err)
	}
	defer reader.Close()

	client, err := b.connect(url)
	if err != nil {
		return nil, err
	}

	//path := strings.TrimPrefix(u.Path, "/")
	dir, name := pathlib.Split(u.Path)
	fmt.Println("PUT", dir, name)
	if dir != "" {
		err := client.ChangeDir(dir)
		if err != nil {
			return nil, fmt.Errorf("ftpStorage: changing directory to %q: %v", dir, err)
		}
		fmt.Println(client.CurrentDir())
	}

	err = client.Stor(name, reader)
	if err != nil {
		return nil, fmt.Errorf("ftpStorage: uploading file for %q: %v", url, err)
	}

	return b.Stat(ctx, url)
}

// Join joins the given URL with the given subpath.
func (b *FTP) Join(url, path string) (string, error) {
	return strings.TrimSuffix(url, "/") + "/" + path, nil
}

// List is not supported by FTP storage.
func (b *FTP) List(ctx context.Context, url string) ([]*Object, error) {
	u, err := urllib.Parse(url)
	if err != nil {
		return nil, fmt.Errorf("ftpStorage: parsing URL: %s", err)
	}

	client, err := b.connect(url)
	if err != nil {
		return nil, err
	}

	resp, err := client.List(u.Path)
	if err != nil {
		return nil, fmt.Errorf("ftpStorage: listing path: %v", err)
	}

	// Special case where the user called List on a regular file.
	if len(resp) == 1 && resp[0].Type == ftp.EntryTypeFile {
		r := resp[0]
		return []*Object{
			{
				URL:          url,
				Name:         strings.TrimPrefix(u.Path, "/"),
				LastModified: r.Time,
				Size:         int64(r.Size),
			},
		}, nil
	}

	// List the objects, recursively.
	var objects []*Object
	for _, r := range resp {
		switch r.Type {

		case ftp.EntryTypeFolder:
			joined, err := b.Join(url, r.Name)
			if err != nil {
				return nil, err
			}

			sub, err := b.List(ctx, joined)
			if err != nil {
				return nil, err
			}
			objects = append(objects, sub...)

		case ftp.EntryTypeLink:
			// Link type is currently not supported. Skip it.
			// TODO there is a "EntryTypeLink" type. can we support that?

		case ftp.EntryTypeFile:
			joined, err := b.Join(url, r.Name)
			if err != nil {
				return nil, err
			}

			obj := &Object{
				URL:          joined,
				Name:         strings.TrimPrefix(pathlib.Join(u.Path, r.Name), "/"),
				LastModified: r.Time,
				Size:         int64(r.Size),
			}
			objects = append(objects, obj)
		}
	}
	return objects, nil
}

// UnsupportedOperations describes which operations (Get, Put, etc) are not
// supported for the given URL.
func (b *FTP) UnsupportedOperations(url string) UnsupportedOperations {
	if err := b.supportsPrefix(url); err != nil {
		return AllUnsupported(err)
	}
	return AllSupported()
}

func (b *FTP) supportsPrefix(url string) error {
	if !strings.HasPrefix(url, "ftp://") && !strings.HasPrefix(url, "sftp://") {
		return &ErrUnsupportedProtocol{"ftpStorage"}
	}
	return nil
}
