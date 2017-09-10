id=`funnel run --cmd 'echo hello world'`
funnel task wait $id
funnel task get $id
