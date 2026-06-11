# PRIDE Checksum Helper

This folder helps you create the `checksum.txt` file required for a PRIDE submission.

It uses a small program to generate SHA-1 checksums in the format expected by PRIDE. For the official background and requirements, see the [PRIDE checksum documentation](https://www.ebi.ac.uk/pride/markdownpage/checksum).

## What's in this folder?

You only need to use one launcher file:

- **Windows**: `run_checksum.bat`
- **macOS / Linux**: `run_checksum.sh`

These launcher files start the checksum program for you. You do not need to open anything inside `bin/` yourself.

### What is `bin/`?

The `bin/` folder contains the actual checksum program, already built and ready to run.

Think of it like this:

- `run_checksum.bat` or `run_checksum.sh` = the button you press
- the file inside `bin/` = the program that does the work

Because different computers need different versions of the program, this folder includes one copy for each common system:

| Folder | For |
| --- | --- |
| `bin/windows-amd64/` | Windows PCs |
| `bin/darwin-arm64/` | Apple Macs with Apple Silicon (M1, M2, M3, etc.) |
| `bin/darwin-amd64/` | Older Intel Macs |
| `bin/linux-amd64/` | Linux computers |

You do not need to install Go, Python, or anything else. As long as you have the full folder, including `bin/`, the launcher will pick the right program automatically.

## Creating `checksum.txt`

Put all the files you want to submit to PRIDE in one folder.

Important:

- All submission files must be directly inside that folder.
- Subfolders are not supported.
- The helper only processes regular files in the top level of the folder.

### Windows

1. Drag your data folder onto `run_checksum.bat`.
2. Wait for the script to finish.
3. A file called `checksum.txt` will be created inside your data folder.

No install step is required.

### macOS / Linux

Run:

```bash
cd "/path/to/checksum script"
./run_checksum.sh "/path/to/your/pride/data/folder"
```

No install step is required.

If macOS says the script is not executable, run:

```bash
chmod +x run_checksum.sh
```

## What Happens During a Run

- The helper prints progress as it processes each file.
- `checksum.txt` is only written at the end, after every file has been hashed successfully.
- While the run is in progress, a temporary file called `checksum.txt.tmp` may appear.
- If a file cannot be read, the helper stops with the file path and error.
- If the run fails, any existing `checksum.txt` is left unchanged.

## Filename Requirements

PRIDE rejects filenames with spaces or unsupported special characters.

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
- Subfolders inside the data folder

## Output Format

The generated file follows the tab-separated SHA-1 format described in the [PRIDE checksum documentation](https://www.ebi.ac.uk/pride/markdownpage/checksum):

```text
# SHA-1 Checksum 
file1.txt	aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d
file2.xml	356a192b7913b04c54574d18c28d46e6395428ab
```

## Troubleshooting

- **Invalid filename**: rename the file to match the requirements above and run again.
- **Unreadable file on a network drive**: check that the mapped drive is still connected, the file is fully copied, and no other program has it open.
- **Missing program file**: make sure you have the full folder, including the `bin/` directory. If files are missing, download the latest copy of this folder, or ask the maintainer to rebuild it.
- **Large files take a long time**: this is normal. Progress is printed while each file is being read.

## For Maintainers

The files in `bin/` are built from the Go source code in `main.go`.

If you change the program, rebuild the ready-to-run copies with:

```bash
./build_binaries.sh
```

Then commit the updated files in `bin/`.

To develop or test from source:

```bash
go test ./...
go run . "/path/to/test/folder"
```
