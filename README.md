# photo-lambda

This code provides the catalog operation for the photo archive.

For CR3 (Canon RAW) and JPEG files, it will try to figure out the timestamp on the file (i.e. when the
photo was taken) and move the incoming files to the archive bucket under prefix/year/month/day/filename

## Building

Assuming go 1.20.5 or better is installed:

```
% go test
% go build
```

## License

Copyright 2022 Little Dog Digital

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this
file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF
ANY KIND, either express or implied. See the License for the specific language
governing permissions and limitations under the License.