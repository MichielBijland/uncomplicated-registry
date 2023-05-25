package version

import "fmt"

var (
	version string = "dev"
	commit  string = "none"
	date    string = "unknown"
)

type MetaData struct {
	Version     string
	Commit      string
	ShortCommit string
	Date        string
}

func (md *MetaData) ShortHash() string {
	if len(md.Commit) >= 7 {
		return md.Commit[:7]
	}
	return md.Commit
}

// String returns a human readable version string.
func (md *MetaData) String() string {
	return fmt.Sprintf("uncomplicated-registry %s (%s) - %s\n", md.Version, md.ShortCommit, md.Date)
}

func GetVersionMetadata() *MetaData {
	return &MetaData{
		Version: version,
		Commit:  commit,
		Date:    date,
	}
}
