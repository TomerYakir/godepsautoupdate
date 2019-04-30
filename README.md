## godepsautoupdate
Script to report on status of dependencies (3rd party libs) - whether there's a newer version/commit available
Works with Go projects that manage the 3rd party libs using `gpm` (Godeps file)

### Usage
1. Clone the repo

2. Switch dir and build the tool
`go build *.go`

3. Output an html report with the results
`./godepsautoupdate --path <GO_DEPS_FILE_PATH> --gopath <GO_PACKAGES_ROOT_PATH>`

Example:
`./godepsautoupdate --path ~/myGoProgram/Godeps --gopath ~/myGoProgram/myroot`

![Report Example](reportScreenshot.png?raw=true "Report Example")

- Clicking on the package link would get to the repo page
- Clicking on New Version would show a git compare between the old and new versions