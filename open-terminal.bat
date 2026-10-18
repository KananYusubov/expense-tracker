@echo off
cd /d "%~dp0"

start http://localhost:8080/ping

wt -d "%CD%\expense-api" --title "api" -p "Command Prompt" cmd /k "go run ./api/main.go" ; new-tab -d "%CD%" --title "pnpm" ; new-tab -d "%CD%" --title "docker" ; new-tab -d "%CD%" --title "git"