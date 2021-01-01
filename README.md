# photo-lambda

This code provides the file-and-resize operation for the photo archive.

## Building

Assuming go 1.15 or better is installed:

```
% go test
% go build
```

## TODO
 1. remove the test that requires a real image on a real bucket
 1. add code to do the copy from source to destination
 1. add code to create the thumbnail