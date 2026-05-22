package core

func ProbeFlags(cfg Config) []string {
	args := AppendNetworkArgs(nil, cfg)
	if cfg.NoPlaylist {
		args = append(args, "--no-playlist")
	}
	return args
}
