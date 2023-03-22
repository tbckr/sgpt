package buildinfo

var (
	version    = "dev"
	commit     = "unset"
	commitDate = "unset"
)

func Version() string {
	return version
}

func Commit() string {
	return commit
}

func CommitDate() string {
	return commitDate
}
