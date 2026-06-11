@echo off
setlocal

set "UV_EXE="

where uv >nul 2>&1
if not errorlevel 1 set "UV_EXE=uv"

if "%UV_EXE%"=="" if exist "%USERPROFILE%\.local\bin\uv.exe" set "UV_EXE=%USERPROFILE%\.local\bin\uv.exe"

if "%UV_EXE%"=="" (
  echo uv is not installed or could not be found.
  echo.
  echo Please double-click install.bat first, then try again.
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

echo Creating PRIDE checksum file for:
echo %DATA_FOLDER%
echo.

"%UV_EXE%" tool run --from pride-checksum pride_checksum --files_dir "%DATA_FOLDER%" --out_path "%DATA_FOLDER%"
if errorlevel 1 (
  echo.
  echo Something went wrong.
  echo.
  echo Common causes:
  echo - uv or pride-checksum is not installed. Run install.bat first.
  echo - One or more filenames contain spaces or special characters.
  echo - The folder contains duplicate filenames.
  echo.
  pause
  exit /b 1
)

echo.
echo Done.
echo checksum.txt was saved in:
echo %DATA_FOLDER%
pause
