package storage

// Google storage (GS)

import (
	"context"
	"fmt"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/util/fsutil"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/storage/v1"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
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
	url := gs.parse(rawurl)

	switch class {
	case File:
		call := gs.svc.Objects.Get(url.bucket, url.path)
		err := download(call, hostPath)
		if err != nil {
			return err
		}
		return nil

	case Directory:
		objects := []*storage.Object{}
		err := gs.svc.Objects.List(url.bucket).Prefix(url.path).Pages(ctx,
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
			call := gs.svc.Objects.Get(url.bucket, obj.Name)
			key := strings.TrimPrefix(obj.Name, url.path)
			err := download(call, path.Join(hostPath, key))
			if err != nil {
				return err
			}
		}
		return nil

	default:
		return fmt.Errorf("Unknown file class: %s", class)
	}
}

func download(call *storage.ObjectsGetCall, hostPath string) error {
	resp, derr := call.Download()
	if derr != nil {
		return derr
	}

	fsutil.EnsurePath(hostPath)
	dest, cerr := os.Create(hostPath)
	if cerr != nil {
		return cerr
	}

	_, werr := io.Copy(dest, resp.Body)
	if werr != nil {
		return werr
	}
	return nil
}

// PutFile copies an object (file) from the host path to GS.
func (gs *GSBackend) PutFile(ctx context.Context, rawurl string, hostPath string) error {
	url := gs.parse(rawurl)

	reader, oerr := os.Open(hostPath)
	if oerr != nil {
		return oerr
	}

	obj := &storage.Object{
		Name: url.path,
	}

	_, err := gs.svc.Objects.Insert(url.bucket, obj).Media(reader).Do()
	return err
}

// SupportsGet indicates whether this backend supports GET storage request.
// For the Google Storage backend, the url must start with "gs://" and the bucket must exist
func (gs *GSBackend) SupportsGet(rawurl string, class tes.FileType) error {
	ok := strings.HasPrefix(rawurl, gsProtocol)
	if !ok {
		return fmt.Errorf("gs: unsupported protocol; expected %s", gsProtocol)
	}
	url := gs.parse(rawurl)
	_, err := gs.svc.Buckets.Get(url.bucket).Do()
	if err != nil {
		return fmt.Errorf("gs: failed to find bucket: %s. error: %v", url.bucket, err)
	}
	return nil
}

// SupportsPut indicates whether this backend supports PUT storage request.
// For the Google Storage backend, the url must start with "gs://" and the bucket must exist
func (gs *GSBackend) SupportsPut(rawurl string, class tes.FileType) error {
	return gs.SupportsGet(rawurl, class)
}

func (gs *GSBackend) parse(rawurl string) *urlparts {
	path := strings.TrimPrefix(rawurl, gsProtocol)
	split := strings.SplitN(path, "/", 2)
	bucket := split[0]
	key := split[1]
	return &urlparts{bucket, key}
}
