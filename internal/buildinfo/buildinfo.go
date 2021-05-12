package buildinfo

var Version = "v0.1.0"
var GitSHA = "000000"
var Server = "http://mini.caprice.run"
var Go = "0.0.0.0.0.0.0"

func All() *map[string]string {
	return &map[string]string{
		"Version": Version,
		"GitSHA":  GitSHA,
		"Server":  Server,
		"Go":      Go,
	}
}
