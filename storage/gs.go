package storage

// Google storage (GS)

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/util/fsutil"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/storage/v1"
)

// The gs url protocol
const gsProtocol = "gs://"

// GSBackend provides access to an GS object store.
type GSBackend struct {
	svc *storage.Service
}

// NewGSBackend creates an GSBackend client instance, give an endpoint URL
// and a set of authentication credentials.
func NewGSBackend(conf config.GSStorage) (*GSBackend, error) {
	ctx := context.Background()
	client := &http.Client{}

	if conf.CredentialsFile != "" {
		// Pull the client configuration (e.g. auth) from a given account file.
		// This is likely downloaded from Google Cloud manually via IAM & Admin > Service accounts.
		bytes, rerr := ioutil.ReadFile(conf.CredentialsFile)
		if rerr != nil {
			return nil, rerr
		}

		config, tserr := google.JWTConfigFromJSON(bytes, storage.CloudPlatformScope)
		if tserr != nil {
			return nil, tserr
		}
		client = config.Client(ctx)
	} else {
		// Pull the information (auth and other config) from the environment,
		// which is useful when this code is running in a Google Compute instance.
		defClient, err := google.DefaultClient(ctx, storage.CloudPlatformScope)
		if err == nil {
			client = defClient
		}
	}

	svc, cerr := storage.New(client)
	if cerr != nil {
		return nil, cerr
	}

	return &GSBackend{svc}, nil
}

// Get copies an object from GS to the host path.
func (gs *GSBackend) Get(ctx context.Context, rawurl string, hostPath string, class tes.FileType) error {
	url, err := gs.parse(rawurl)
	if err != nil {
		return err
	}

	switch class {
	case File:
		err := fsutil.EnsurePath(hostPath)
		if err != nil {
			return err
		}

		call := gs.svc.Objects.Get(url.bucket, url.path)
		return download(ctx, call, hostPath)

	case Directory:
		err := fsutil.EnsureDir(hostPath)
		if err != nil {
			return err
		}

		objects := []*storage.Object{}
		err = gs.svc.Objects.List(url.bucket).Prefix(url.path).Pages(ctx,
			func(objs *storage.Objects) error {
				objects = append(objects, objs.Items...)
				return nil
			})
		if err != nil {
			return err
		}
		if len(objects) == 0 {
			return ErrEmptyDirectory
		}

		for _, obj := range objects {
			if strings.HasSuffix(obj.Name, "/") {
				continue
			}
			call := gs.svc.Objects.Get(url.bucket, obj.Name)
			key := strings.TrimPrefix(obj.Name, url.path)
			err = download(ctx, call, path.Join(hostPath, key))
			if err != nil {
				return err
			}
		}
		return nil

	default:
		return fmt.Errorf("Unknown file class: %s", class)
	}
}

func download(ctx context.Context, call *storage.ObjectsGetCall, hostPath string) (err error) {
	resp, err := call.Download()
	if err != nil {
		return err
	}

	err = fsutil.EnsurePath(hostPath)
	if err != nil {
		return err
	}
	dest, err := os.Create(hostPath)
	if err != nil {
		return err
	}
	defer func() {
		cerr := dest.Close()
		if cerr != nil {
			err = fmt.Errorf("%v; %v", err, cerr)
		}
	}()

	_, err = io.Copy(dest, fsutil.Reader(ctx, resp.Body))
	return err
}

// PutFile copies an object (file) from the host path to GS.
func (gs *GSBackend) PutFile(ctx context.Context, rawurl string, hostPath string) error {
	url, err := gs.parse(rawurl)
	if err != nil {
		return err
	}

	reader, err := os.Open(hostPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	obj := &storage.Object{
		Name: url.path,
	}

	_, err = gs.svc.Objects.Insert(url.bucket, obj).Media(fsutil.Reader(ctx, reader)).Do()
	return err
}

// SupportsGet indicates whether this backend supports GET storage request.
// For the Google Storage backend, the url must start with "gs://" and the bucket must exist
func (gs *GSBackend) SupportsGet(rawurl string, class tes.FileType) error {
	url, err := gs.parse(rawurl)
	if err != nil {
		return err
	}
	_, err = gs.svc.Buckets.Get(url.bucket).Do()
	if err != nil {
		return fmt.Errorf("googleStorage: failed to find bucket: %s. error: %v", url.bucket, err)
	}
	return nil
}

// SupportsPut indicates whether this backend supports PUT storage request.
// For the Google Storage backend, the url must start with "gs://" and the bucket must exist
func (gs *GSBackend) SupportsPut(rawurl string, class tes.FileType) error {
	return gs.SupportsGet(rawurl, class)
}

func (gs *GSBackend) parse(rawurl string) (*urlparts, error) {
	if !strings.HasPrefix(rawurl, gsProtocol) {
		return nil, &ErrUnsupportedProtocol{"googleStorage"}
	}

	path := strings.TrimPrefix(rawurl, gsProtocol)
	if path == "" {
		return nil, &ErrInvalidURL{"googleStorage"}
	}

	split := strings.SplitN(path, "/", 2)
	url := &urlparts{}
	if len(split) > 0 {
		url.bucket = split[0]
	}
	if len(split) == 2 {
		url.path = split[1]
	}
	return url, nil
}
