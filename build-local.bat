@echo off
cd /d C:\Users\User\Documents\Trae\nodepath-chat-1
set CGO_ENABLED=0
echo Building server without CGO...
go build -o bin\server.exe cmd\server\main.go
echo Build completed!
