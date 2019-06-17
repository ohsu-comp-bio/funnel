FROM docker:18.09.6
ADD funnel /opt/funnel/funnel
VOLUME /opt/funnel/funnel-work-dir
EXPOSE 8000 9090
ENTRYPOINT ["/opt/funnel/funnel"]
