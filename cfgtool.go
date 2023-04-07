
func readConfig() {
	var cfg configfile

	f := readFile("config.toml")

	err := toml.Unmarshal(f, &cfg)
	if err != nil {
		panic(err)
	}
	attrPath := cfg.Attributes.Path
	fmt.Println("path: ", attrPath)
	// return cfg
}

func readFile(filename string) []byte {
	content, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	return content
}