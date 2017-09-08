version=`git describe --tags --long --dirty`
commit=`git rev-parse --short HEAD`
branch=`git symbolic-ref -q --short HEAD`
dt=`date`

# Write out the package.
cat << EOF > data.go
package version

// Build and version details
const (
	GitCommit = "$commit"
	GitBranch = "$branch"
	BuildDate = "$dt"
	Version   = "$version"
)
EOF

