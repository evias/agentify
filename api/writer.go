package api

import (
	"io/ioutil"
	"strings"
)

// WriteAgentsFile renders and writes the AGENTS.md content.
func WriteAgentsFile(dest string, rs *RepoScan) error {
	content := &strings.Builder{}
	content.WriteString(rs.Result)

	str := content.String()
	return ioutil.WriteFile(dest, []byte(str), 0644)
}
