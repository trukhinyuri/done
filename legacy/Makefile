go builall:
	env GOOS=darwin GOARCH=amd64 go build -o bin/done-darwin-amd64 done.go
	env GOOS=linux GOARCH=amd64 go build -o bin/done-linux-amd64 done.go
	env GOOS=windows GOARCH=amd64 go build -o bin/done-windows-amd64.exe done.go
	env GOOS=windows GOARCH=386 go build -o bin/done-windows-386.exe done.go
	chmod +x bin/done-darwin-amd64
	chmod +x bin/done-linux-amd64
	chmod +x bin/done-windows-amd64.exe
	chmod +x bin/done-windows-386.exe
	- yes | -cp bin/done-darwin-amd64 ~/Dropbox/done_work/done
	- yes | -cp bin/done-windows-386.exe ~/Dropbox/don./done_work/done.exe
	- yes | cp -r frontend ~/Dropbox/done_work/frontend/
	env GOOS=darwin GOARCH=amd64 go build -o done done.go
