package storage

// Google storage (GS)

import (
	"context"
	"fmt"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/storage/v1"
	"io"
	"io/ioutil"
	"net/http"
	urllib "net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"tes/config"
	"tes/util"
)

// The gs url protocol
const gsscheme = "gs"

// GSBackend provides access to an GS object store.
type GSBackend struct {
	svc *storage.Service
}

// NewGSBackend creates an GSBackend client instance, give an endpoint URL
// and a set of authentication credentials.
func NewGSBackend(conf config.GSStorage) (*GSBackend, error) {
	ctx := context.Background()
	client := &http.Client{}

	if conf.AccountFile != "" {
		// Pull the client configuration (e.g. auth) from a given account file.
		// This is likely downloaded from Google Cloud manually via IAM & Admin > Service accounts.
		bytes, rerr := ioutil.ReadFile(conf.AccountFile)
		if rerr != nil {
			return nil, rerr
		}

		config, tserr := google.JWTConfigFromJSON(bytes, storage.CloudPlatformScope)
		if tserr != nil {
			return nil, tserr
		}
		client = config.Client(ctx)
	} else if conf.FromEnv {
		// Pull the information (auth and other config) from the environment,
		// which is useful when this code is running in a Google Compute instance.
		var err error
		client, err = google.DefaultClient(ctx, storage.CloudPlatformScope)
		if err != nil {
			log.Error("Error connecting Google Storage client. Defaulting to anonymous.", err)
			// No auth config could be found, so default to anonymous.
		}
	}

	svc, cerr := storage.New(client)
	if cerr != nil {
		return nil, cerr
	}

	return &GSBackend{svc}, nil
}

// Get copies an object from GS to the host path.
func (gs *GSBackend) Get(ctx context.Context, rawurl string, hostPath string, class string) error {
	log.Info("Starting download", "url", rawurl)

	url, perr := parse(rawurl)
	if perr != nil {
		return perr
	}

	if class == File {
		call := gs.svc.Objects.Get(url.bucket, url.path)
		err := download(call, hostPath)
		if err != nil {
			return err
		}
		log.Info("Finished file download", "url", rawurl, "hostPath", hostPath)
		return nil

	} else if class == Directory {
		// TODO not handling pagination
		objects, _ := gs.svc.Objects.List(url.bucket).Prefix(url.path).Do()
		for _, obj := range objects.Items {
			call := gs.svc.Objects.Get(url.bucket, obj.Name)
			err := download(call, path.Join(hostPath, obj.Name))
			if err != nil {
				return err
			}
		}
		log.Info("Finished directory download", "url", rawurl, "hostPath", hostPath)
		return nil
	}
	return fmt.Errorf("Unknown file class: %s", class)
}

func download(call *storage.ObjectsGetCall, hostPath string) error {
	resp, derr := call.Download()
	if derr != nil {
		return derr
	}

	util.EnsurePath(hostPath)
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

// Put copies an object (file) from the host path to GS.
func (gs *GSBackend) Put(ctx context.Context, rawurl string, hostPath string, class string) error {
	log.Info("Starting upload", "url", rawurl)

	url, perr := parse(rawurl)
	if perr != nil {
		return perr
	}

	if class == File {
		reader, oerr := os.Open(hostPath)
		if oerr != nil {
			return oerr
		}

		obj := &storage.Object{
			Name: url.path,
		}

		_, err := gs.svc.Objects.Insert(url.bucket, obj).Media(reader).Do()
		if err != nil {
			return err
		}
		return nil

	} else if class == Directory {
		err := filepath.Walk(hostPath, func(p string, f os.FileInfo, err error) error {
			if !f.IsDir() {
				// TODO do what?
				rel, _ := filepath.Rel(hostPath, p)
				// TODO do what?
				gs.Put(ctx, rawurl+"/"+rel, p, File)
				log.Debug("Subpath", "full", p, "rel", rel)
			}
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	}
	log.Info("Finished upload", "url", rawurl, "hostPath", hostPath)
	return fmt.Errorf("Unknown file class: %s", class)
}

// Supports returns true if this backend supports the given storage request.
// The Google Storage backend supports URLs which have a "gs://" scheme.
func (gs *GSBackend) Supports(rawurl string, hostPath string, class string) bool {
	_, err := parse(rawurl)
	if err != nil {
		return false
	}
	return true
}

type urlparts struct {
	bucket string
	path   string
}

func parse(rawurl string) (*urlparts, error) {
	url, err := urllib.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	if url.Scheme != gsscheme {
		return nil, fmt.Errorf("Invalid URL scheme '%s' for Google Storage backend in url: %s", url.Scheme, rawurl)
	}

	bucket := url.Host
	path := strings.TrimLeft(url.EscapedPath(), "/")
	return &urlparts{bucket, path}, nil
}
