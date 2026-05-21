package core

func ConfigForProfile(cfg Config) Config {
	out := cfg
	out.URL = ""
	out.BatchURLs = ""
	out.LoadInfoJSON = ""
	return out
}

func ApplyProfileConfig(dst *Config, profile Config) {
	url := dst.URL
	batch := dst.BatchURLs
	loadJSON := dst.LoadInfoJSON
	*dst = profile
	dst.URL = url
	dst.BatchURLs = batch
	dst.LoadInfoJSON = loadJSON
}
