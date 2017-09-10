# Test that the streaming logs pick up a single character.
# This ensures that the streaming works even when a small
# amount of logs are written (which was once a bug).
id=`funnel run 'echo a'`
funnel task wait $id
funnel task get --view FULL $id
