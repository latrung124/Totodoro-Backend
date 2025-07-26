@echo off
REM Build script for Windows
SET GOOS=windows
SET GOARCH=amd64
go build -o bin\totodoro-backend.exe main.go