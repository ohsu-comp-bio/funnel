# https://github.com/stilliard/docker-pure-ftpd
FROM stilliard/pure-ftpd
RUN useradd -d /var/ftp -u 2001 -s /sbin/nologin ftp
RUN mkdir /var/ftp
CMD /run.sh -c 30 -C 10 -l puredb:/etc/pure-ftpd/pureftpd.pdb -j -R -P localhost -p 30000:30059
