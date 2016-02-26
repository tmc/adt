// Command
package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
)

var (
	flagPath       = flag.String("path", ".", "path to ADT files")
	flagVerbose    = flag.Bool("v", false, "verbose")
	flagAddr       = flag.String("http", ":7001", "listen address")
	flagConfig     = flag.String("conf", "", "path to config json")
	flagPublicKey  = flag.String("tlscrt", "", "path to tls certificate")
	flagPrivateKey = flag.String("tlskey", "", "path to tls private key")
)

func main() {
	flag.Parse()

	cfg, err := loadConfig(*flagConfig)
	if err != nil {
		log.Fatalln(err)
	}
	srv := NewADTHTTPServer(cfg, *flagPath, *flagVerbose)
	if err := srv.Serve(*flagAddr, *flagPublicKey, *flagPrivateKey); err != nil {
		log.Fatalln(err)
	}
}

func loadConfig(path string) (Config, error) {
	if path == "" {
		return nil, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	result := Config{}
	return result, json.NewDecoder(f).Decode(&result)
}
