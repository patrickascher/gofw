package json_test

import (
	"github.com/patrickascher/gofw/config/json"
	"testing"
)

// BenchmarkJson_Parse is testing a normal config and a config with an existing env config file (two file reads).
func BenchmarkJson_Parse(b *testing.B) {
	mockFiles()
	defer removeMockFiles()

	jc := json.New()
	b.ResetTimer()

	cases := []struct {
		Env      string
		Filepath string
	}{
		{Env: "", Filepath: "config.json"},
		{Env: "dev", Filepath: "config.json"},
	}
	for _, bc := range cases {
		b.Run(bc.Env, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				conf := &mockConfig{}
				_ = jc.Parse(conf, bc.Env, json.Options{Filepath: bc.Filepath})
			}
		})
	}
}
