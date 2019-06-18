## godepsautoupdate
Script to report on status of dependencies (3rd party libs) - whether there's a newer version/commit available.
Works with Go projects that manage the 3rd party libs using the following dependecy file formats:
1. `gpm` (Godeps file, default)
2. `go dep` (Gopkg file)
3. `go modules` (coming soon!)

### Usage
1. Get the tool
```
git clone https://github.com/TomerYakir/godepsautoupdate.git
cd godepsautoupdate
export GOPATH=`pwd`
go build godepsautoupdate.go
```

2. Output an html report with the results
```
cd bin
./godepsautoupdate --path <GO_DEPS_FILE_PATH> --gopath <GO_PACKAGES_ROOT_PATH>
```

Example - gpm format:
```
cd bin
./godepsautoupdate --path ~/myGoProgram/Godeps --gopath ~/myGoProgram/myroot
```

Example #2 - dep (gopkg) format:
```
cd bin
./godepsautoupdate --path ~/myGoProgram/Gopkg.toml --gopath ~/myGoProgram/myroot --deptype dep
```


![Report Example](reportScreenshot.png?raw=true "Report Example")

- Clicking on the package link would get to the repo page
- Clicking on New Version would show a git compare between the old and new versions

3. Update the dependency file
```
cd bin
./godepsautoupdate --path ~/myGoProgram/Godeps --gopath ~/myGoProgram/myroot --updateFile
```

### Developer notes
If the reportTemplate.html changes, generate the bin data using `go-bindata -func GetHtmlTemplateBinData reportTemplate.html`.