id=`funnel run 'echo hello'`
funnel task wait $id
funnel task cancel $id
funnel task get --view MINIMAL $id
