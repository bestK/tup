# Telegraph Anonymous Uploader

This is a tool for anonymously uploading files to telegra.ph.

## Installation

To install the tool, you can download the binary from the releases page, or build it from source using Go:

```bash
go build -o tup main.go
```

## Usage

To use the tool, you can run it with the following command-line arguments:

``` shell
./tup -h

-input string
    Path to the file to upload,if > 5MB,split it
-output string
    Directory to output the parts to
-size int
    Size of each part in MB (default 4)
-upload-url string
    The url to upload parts to (default "https://telegra.ph/upload")

```


For example:
``` shell
./tup -input path/to/file  

ls -l
<filename>.json
```

## TODO
-  Merge split files
