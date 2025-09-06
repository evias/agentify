package api

type RepoScan struct {
	Root   string // local filesystem path
	Readme string // raw README.md contents (if found)
	Result string // llm model response text
}
