package version

import "fmt"

var (
	GitTag    = ""
	GitCommit = ""
	GitBranch = ""
	BuildTime = ""
	GoVersion = ""
)

// FullVersion show the version info
func FullVersion() string {
	version := fmt.Sprintf("Version   : %s\nBuild Time: %s\nGit Branch: %s\nGit Commit: %s\nGo Version: %s\n", GitTag, BuildTime, GitBranch, GitCommit, GoVersion)
	return version
}

// Short 版本缩写
func Short() string {
	return fmt.Sprintf("%s[%s %s]", GitTag, BuildTime, GitCommit)
}
