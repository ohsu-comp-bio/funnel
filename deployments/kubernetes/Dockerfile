FROM docker:stable-dind
ADD funnel /usr/bin/funnel
ADD *.json /opt/funnel/examples/
ADD entrypoint.sh /usr/local/bin/entrypoint.sh

ENTRYPOINT ["entrypoint.sh"]
