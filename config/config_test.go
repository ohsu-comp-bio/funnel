package config

import "testing"

func TestNodeResourceConfigParsing(t *testing.T) {
	yaml := `
Scheduler:
  Node:
    Resources:
      Cpus: 42
      RamGb: 2.5
      DiskGb: 50.0
`
	conf := Config{}
	Parse([]byte(yaml), &conf)

	if conf.Scheduler.Node.Resources.Cpus != 42 {
		t.Fatal("unexpected cpus")
	}
	if conf.Scheduler.Node.Resources.RamGb != 2.5 {
		t.Fatal("unexpected ram")
	}
	if conf.Scheduler.Node.Resources.DiskGb != 50.0 {
		t.Fatal("unexpected disk")
	}
}

func TestEnsureServerProperties(t *testing.T) {
	serverAddress := "test:9090"

	conf := Config{}
	conf.Server.HostName = "test"
	conf.Server.RPCPort = "9090"
	result := EnsureServerProperties(conf)

	if result.Server.RPCAddress() != serverAddress {
		t.Fatal("unexpected server address")
	}
	if result.Scheduler.Node.ServerAddress != serverAddress {
		t.Fatal("unexpected node server address")
	}
	if result.Worker.EventWriters.RPC.ServerAddress != serverAddress {
		t.Fatal("unexpected server address in worker config")
	}
	if result.Worker.TaskReaders.RPC.ServerAddress != serverAddress {
		t.Fatal("unexpected server address in worker config")
	}

	conf = Config{}
	conf.Scheduler.Node.ServerAddress = serverAddress
	result = EnsureServerProperties(conf)

	if result.Server.RPCAddress() != serverAddress {
		t.Fatal("unexpected server address")
	}
	if result.Scheduler.Node.ServerAddress != serverAddress {
		t.Fatal("unexpected node server address")
	}
	if result.Worker.TaskReaders.RPC.ServerAddress != serverAddress {
		t.Fatal("unexpected server address in worker task reader config")
	}
	if result.Worker.EventWriters.RPC.ServerAddress != serverAddress {
		t.Fatal("unexpected server address in worker event writer config")
	}

	conf = Config{}
	conf.Worker.EventWriters.RPC.ServerAddress = serverAddress
	result = EnsureServerProperties(conf)

	if result.Server.RPCAddress() != serverAddress {
		t.Fatal("unexpected server address")
	}
	if result.Scheduler.Node.ServerAddress != serverAddress {
		t.Fatal("unexpected node server address")
	}
	if result.Worker.TaskReaders.RPC.ServerAddress != serverAddress {
		t.Fatal("unexpected server address in worker task reader config")
	}
	if result.Worker.EventWriters.RPC.ServerAddress != serverAddress {
		t.Fatal("unexpected server address in worker event writer config")
	}

	conf = Config{}
	conf.Worker.TaskReaders.RPC.ServerAddress = serverAddress
	result = EnsureServerProperties(conf)

	if result.Server.RPCAddress() != serverAddress {
		t.Fatal("unexpected server address")
	}
	if result.Scheduler.Node.ServerAddress != serverAddress {
		t.Fatal("unexpected node server address")
	}
	if result.Worker.TaskReaders.RPC.ServerAddress != serverAddress {
		t.Fatal("unexpected server address in worker task reader config")
	}
	if result.Worker.EventWriters.RPC.ServerAddress != serverAddress {
		t.Fatal("unexpected server address in worker event writer config")
	}

	// Test overwrite order
	// Expected precedence:
	// Server > Node > TaskReader > EventWriter

	conf = Config{}
	serverAddress = "server:9999"

	conf.Worker.EventWriters.RPC.ServerAddress = "eventwriter:9999"
	conf.Worker.TaskReaders.RPC.ServerAddress = "taskreader:9999"
	conf.Scheduler.Node.ServerAddress = "node:9999"
	conf.Server.HostName = "server"
	conf.Server.RPCPort = "9999"

	result = EnsureServerProperties(conf)

	if result.Server.RPCAddress() != serverAddress {
		t.Fatal("unexpected server address")
	}
	if result.Scheduler.Node.ServerAddress != serverAddress {
		t.Fatal("unexpected node server address")
	}
	if result.Worker.TaskReaders.RPC.ServerAddress != serverAddress {
		t.Fatal("unexpected server address in worker task reader config")
	}
	if result.Worker.EventWriters.RPC.ServerAddress != serverAddress {
		t.Fatal("unexpected server address in worker event writer config")
	}

	conf = Config{}
	serverAddress = "node:9999"

	conf.Worker.EventWriters.RPC.ServerAddress = "eventwriter:9999"
	conf.Worker.TaskReaders.RPC.ServerAddress = "taskreader:9999"
	conf.Scheduler.Node.ServerAddress = "node:9999"

	result = EnsureServerProperties(conf)

	if result.Server.RPCAddress() != serverAddress {
		t.Fatal("unexpected server address")
	}
	if result.Scheduler.Node.ServerAddress != serverAddress {
		t.Fatal("unexpected node server address")
	}
	if result.Worker.TaskReaders.RPC.ServerAddress != serverAddress {
		t.Fatal("unexpected server address in worker task reader config")
	}
	if result.Worker.EventWriters.RPC.ServerAddress != serverAddress {
		t.Fatal("unexpected server address in worker event writer config")
	}

	conf = Config{}
	serverAddress = "taskreader:9999"

	conf.Worker.EventWriters.RPC.ServerAddress = "eventwriter:9999"
	conf.Worker.TaskReaders.RPC.ServerAddress = "taskreader:9999"

	result = EnsureServerProperties(conf)

	if result.Server.RPCAddress() != serverAddress {
		t.Fatal("unexpected server address")
	}
	if result.Scheduler.Node.ServerAddress != serverAddress {
		t.Fatal("unexpected node server address")
	}
	if result.Worker.TaskReaders.RPC.ServerAddress != serverAddress {
		t.Fatal("unexpected server address in worker task reader config")
	}
	if result.Worker.EventWriters.RPC.ServerAddress != serverAddress {
		t.Fatal("unexpected server address in worker event writer config")
	}
}
