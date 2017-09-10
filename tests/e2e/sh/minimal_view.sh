id=`funnel run --cmd 'echo hello world' --name 'foo' --contents in=./testdata/hello.txt`
funnel task wait $id
funnel task get --view MINIMAL $id
