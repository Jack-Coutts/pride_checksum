# PRIDE Checksum Helper

This folder helps you create the `checksum.txt` file required for a PRIDE submission.

It uses `uv` to run the official PRIDE checksum tool, so you do not need to manually install Python packages.

## First-Time Setup

1. Double-click `install.bat`.
2. If Windows asks whether the file can run, allow it.
3. Wait until it says installation is complete.

You only need to do this once.

`install.bat` installs `uv` if needed, then prepares the official `pride-checksum` tool.

## Creating `checksum.txt`

1. Put all the files you want to submit to PRIDE in one folder.
2. Make sure the filenames follow the PRIDE checksum tool requirements below.
3. Drag your data folder onto `run_checksum.bat`.
4. Wait for the script to finish.
5. A file called `checksum.txt` will be created inside your data folder.

## Filename Requirements

The official PRIDE checksum tool rejects filenames with spaces or unsupported special characters.

Use filenames with letters, numbers, underscores, hyphens, and normal file extensions.

Good examples:

```text
sample_01.raw
experiment-02.mzML
results_file.txt
```

Bad examples:

```text
sample 01.raw
experiment@02.mzML
.hidden_file
```

Also avoid:

- Hidden files starting with `.`
- Duplicate filenames, even if they are in different folders

## Notes

- The first run can take a while for large files.
- If you run it again later, unchanged files may be reused, so it can be faster.
- If the script reports an invalid filename, rename the file to match the requirements above and run it again.
- If the script says `uv` or `pride-checksum` is not installed, run `install.bat`.
