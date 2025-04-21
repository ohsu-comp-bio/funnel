> [!WARNING]
> Files in this directory are in active development and will be moved to a separate public repository (e.g. https://github.com/ohsu-comp-bio/funnel-example-plugin) in the near future 🔮

# Updating the Protocol

> [!NOTE]
> Adapted from [gRPC go-plugin example](https://github.com/hashicorp/go-plugin/tree/main/examples/grpc#updating-the-protocol) from Hashicorp

Updates to the protocol are managed by [`buf`](https://github.com/bufbuild/buf)

## 1. Install `buf`

```sh
brew install bufbuild/buf/buf
```

## 2. Update Protocol

> [!TIP]
> This step uses the [`buf.gen.yaml`](./buf.gen.yaml) as the configuration file

```sh
📖 Reading protobuf schema:
- proto/auth.proto

🔧 Generating code with buf...

✅ Verifying expected files were generated...

🎉 Go and Python Protobuf/gRPC files generated:
- proto/auth.pb.go
- proto/auth_grpc.pb.go
- plugin-python/proto/auth_pb2.py
- plugin-python/proto/auth_pb2_grpc.py
```

## 3. Verify Output

If all goes well `buf generate` will read in the `proto/auth.proto` schema definition and output four files split between the `plugin-go` and `plugin-python` directories:

```sh
./plugins
├── plugin-go
│   ├── auth_impl.go
│   ├── auth_impl_test.go
│   └── proto
│       ├── auth.pb.go         <---- 1) Protobuf Message Types (Go)
│       └── auth_grpc.pb.go    <---- 2) gRPC Service Interface (Go)
├── plugin-python
│   └── proto
│       ├── auth_pb2.py        <---- 3) Protobuf Message Types (Python)
│       └── auth_pb2_grpc.py   <---- 4) gRPC Service Interface (Python)
└── proto
    ├── auth.pb.go
    ├── auth.proto
    └── auth_grpc.pb.go
```

# Additional Resources

- https://buf.build/docs/generate/tutorial/
- https://github.com/hashicorp/go-plugin/tree/main/examples/grpc
- https://grpc.io/
- https://protobuf.dev/
