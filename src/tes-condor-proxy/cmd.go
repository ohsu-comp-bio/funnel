package main

import (
	"flag"
	"os"
	"path"
	"path/filepath"
	"tes/autoscaler"
)

func main() {
	// Get the directory this executable is in
	thisdir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}

	// Define a default path for the tes-worker binary
	binPath := path.Join(thisdir, "tes-worker")

	portArg := flag.String("port", "9054", "Port to listen on")
	binArg := flag.String("bin", binPath, "Path of the tes-worker binary")
	flag.Parse()

	pxy := autoscaler.NewCondorProxy(*binArg)
	// TODO validate port arg? should be int type instead?
	pxy.Start(*portArg)
}
