SET CGO_ENABLED=0
SET GOOS=windows
SET GOARCH=amd64
go build -o windows_inotify_produce.exe

7z a bin/inotify_produce.zip ./

