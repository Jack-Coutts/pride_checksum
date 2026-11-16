@echo off
setlocal

rem No %VAR% inside ( ) blocks: a ")" in a path, as in "folder (1)", ends the block early and cmd exits before pause.
set "HELPER_EXE=%~dp0bin\windows-amd64\pride-checksum-helper.exe"
set "DATA_FOLDER=%~1"

if not exist "%HELPER_EXE%" goto missing_helper
if "%DATA_FOLDER%"=="" goto no_folder
if not exist "%DATA_FOLDER%\" goto not_a_folder

"%HELPER_EXE%" "%DATA_FOLDER%"
if errorlevel 1 goto helper_failed

echo.
pause
exit /b 0

:missing_helper
echo pride-checksum-helper.exe is missing from:
echo "%HELPER_EXE%"
echo.
echo Please download the latest version of this folder, or ask the maintainer to run build_binaries.sh.
pause
exit /b 1

:no_folder
echo Drag your data folder onto this file.
echo.
echo A file called checksum.txt will be created inside that folder.
pause
exit /b 1

:not_a_folder
echo This does not look like a folder:
echo "%DATA_FOLDER%"
echo.
echo Please drag the folder containing your PRIDE submission files onto this file.
pause
exit /b 1

:helper_failed
echo.
echo Something went wrong.
echo.
echo Common causes:
echo - One or more filenames contain spaces or special characters.
echo - The folder contains subfolders or unreadable files.
echo.
pause
exit /b 1
