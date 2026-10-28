@echo off
setlocal
set "APP=%~dp0web\index.html"
set "URL=file:///%APP:\=/%"
set "PROFILE=%LOCALAPPDATA%\MicroMD\profile"

set "EDGE=%ProgramFiles(x86)%\Microsoft\Edge\Application\msedge.exe"
if not exist "%EDGE%" set "EDGE=%ProgramFiles%\Microsoft\Edge\Application\msedge.exe"
if not exist "%EDGE%" set "EDGE=%LOCALAPPDATA%\Microsoft\Edge\Application\msedge.exe"

set "CHROME=%ProgramFiles%\Google\Chrome\Application\chrome.exe"
if not exist "%CHROME%" set "CHROME=%ProgramFiles(x86)%\Google\Chrome\Application\chrome.exe"
if not exist "%CHROME%" set "CHROME=%LOCALAPPDATA%\Google\Chrome\Application\chrome.exe"

if exist "%EDGE%" call :run "%EDGE%"
if exist "%EDGE%" goto :eof
if exist "%CHROME%" call :run "%CHROME%"
if exist "%CHROME%" goto :eof
start "" "%APP%"
goto :eof

:run
start "" %1 --app="%URL%" --user-data-dir="%PROFILE%" --window-size=1280,860 --no-first-run --no-default-browser-check --disable-extensions --disable-sync --disable-background-networking --disable-component-update --disable-default-apps --no-service-autorun --disable-features=Translate,MediaRouter --disk-cache-size=20971520
exit /b
