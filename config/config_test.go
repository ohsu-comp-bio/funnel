package config

import "testing"

func TestWorkerResourceConfigParsing(t *testing.T) {
	yaml := `
Worker:
  Resources:
    Cpus: 42
    RamGb: 2.5
    DiskGb: 50.0
  `
	conf := Config{}
	Parse([]byte(yaml), &conf)

	if conf.Worker.Resources.Cpus != 42 {
		t.Fatal("unexpected cpus")
	}
	if conf.Worker.Resources.RamGb != 2.5 {
		t.Fatal("unexpected ram")
	}
	if conf.Worker.Resources.DiskGb != 50.0 {
		t.Fatal("unexpected disk")
	}
}
