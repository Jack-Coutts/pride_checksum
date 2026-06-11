@echo off
setlocal

echo Installing uv and the PRIDE checksum tool...
echo.

set "UV_EXE="

where uv >nul 2>&1
if not errorlevel 1 set "UV_EXE=uv"

if "%UV_EXE%"=="" if exist "%USERPROFILE%\.local\bin\uv.exe" set "UV_EXE=%USERPROFILE%\.local\bin\uv.exe"

if "%UV_EXE%"=="" (
  echo uv was not found. Installing uv...
  echo.
  powershell -ExecutionPolicy ByPass -NoProfile -Command "irm https://astral.sh/uv/install.ps1 | iex"
  if errorlevel 1 (
    echo.
    echo Could not install uv.
    echo Please check your internet connection and try again.
    pause
    exit /b 1
  )
)

if "%UV_EXE%"=="" if exist "%USERPROFILE%\.local\bin\uv.exe" set "UV_EXE=%USERPROFILE%\.local\bin\uv.exe"

if "%UV_EXE%"=="" (
  echo.
  echo uv may have installed successfully, but this window cannot find it yet.
  echo Close this window, double-click install.bat again, and then continue.
  pause
  exit /b 1
)

"%UV_EXE%" --version
if errorlevel 1 (
  echo.
  echo uv was found but did not run correctly.
  pause
  exit /b 1
)

"%UV_EXE%" tool run --from pride-checksum pride_checksum --help >nul
if errorlevel 1 (
  echo.
  echo Could not prepare pride-checksum.
  echo Please check your internet connection and try again.
  pause
  exit /b 1
)

echo.
echo Installation complete.
echo You can now drag your data folder onto run_checksum.bat.
pause
