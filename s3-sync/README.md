<h2>s3-sync</h2>
Golang utility to watch the specified local directory for changes and sync the changes with the specified AWS S3 directory.
<br></br>
<h3>How to run</h3>

- Clone the repo: 

```
gh repo clone hassaanakram/goto-tools
```
- Navigate to the project directory:

```
goto-tools/s3-sync
```
- Run the following command to build binary:

```
go build -o s3-sync
```
- Run the binary as following:

```
./s3-sync --dir </Path/to/local/dir> --s3_url <s3_url/of/remote/directory>
```

<h3>Requirements</h3>

- go 1.19
- AWS account credentials should be setup in the ~/.aws/config file or exported to terminal. The utility currently does not explicitly report invalid or missing credentials. 

<h3>Supported actions</h3>

Currently, the utility registers the following operations to sync:
- New File creation
- File creation by copy
- File modification

The utility does NOT sync the following operations:
- File rename
- File deletion
