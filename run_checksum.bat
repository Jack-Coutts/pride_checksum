@echo off
setlocal

set "HELPER_EXE=%~dp0bin\windows-amd64\pride-checksum-helper.exe"

if not exist "%HELPER_EXE%" (
  echo pride-checksum-helper.exe is missing from:
  echo %HELPER_EXE%
  echo.
  echo Please download the latest version of this folder, or ask the maintainer to run build_binaries.sh.
  pause
  exit /b 1
)

if "%~1"=="" (
  echo Drag your data folder onto this file.
  echo.
  echo A file called checksum.txt will be created inside that folder.
  pause
  exit /b 1
)

set "DATA_FOLDER=%~1"

if not exist "%DATA_FOLDER%\" (
  echo This does not look like a folder:
  echo %DATA_FOLDER%
  echo.
  echo Please drag the folder containing your PRIDE submission files onto this file.
  pause
  exit /b 1
)

"%HELPER_EXE%" "%DATA_FOLDER%"
if errorlevel 1 (
  echo.
  echo Something went wrong.
  echo.
  echo Common causes:
  echo - One or more filenames contain spaces or special characters.
  echo - The folder contains subfolders or unreadable files.
  echo.
  pause
  exit /b 1
)

echo.
pause
