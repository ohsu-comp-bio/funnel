# This starts a docker container in detached mode, which runs an FTP server.
# This expects to be run in this directory.
# Data and users have been pre-configured.
docker run --name ftp --rm -d -p 21:21 -p 30000-30009:30000-30009 -v `pwd`/_scratch/ftp-test/:/home/ftpusers -v `pwd`/_scratch/ftp-var:/var/ftp -v `pwd`/_scratch/passwd:/etc/pure-ftpd/passwd funnel-ftp-test-server
