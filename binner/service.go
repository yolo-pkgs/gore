package binner

const goProxyURL = "https://proxy.golang.org"

type Bin struct {
	Binary      string
	Path        string
	Mod         string
	ModVersion  string
	LastVersion string
	Updatable   string
}

type Binner struct {
	Bins
}
