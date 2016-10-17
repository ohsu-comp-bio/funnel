
DEVELOPMENT
-----------

This reference server is built in GO, serving the protocol described in 
https://github.com/ga4gh/task-execution-schemas (note, the repo is linked as a
git submodule, so make sure to use `git clone --recursive`)

The code structure is based on two binary programs

1) tes-server: Provide client API endpoint, web UI, manages work queue.
The client facing HTTP port (by default 8000) uses a GO web router to serve static HTML
elements and proxy API requests to a GRPC endpoint, using the GRPC gateway code 
(https://github.com/grpc-ecosystem/grpc-gateway)
The GRPC endpoint (usually bound to port 9090) is based on the auto-generated GO 
protoc GRPC server. It is bound to two services: the GA4GH Task Execution Service 
and a internal service protocol that is used by the workers to access the work queue.

2) tes-worker: A client that uses the internal service protocol to contact task 
scheduler (on port 9090) and requests jobs. It is then responsible for obtaining the 
required input files (thus it much have code and credentials to access object store),
run the docker container with the provided arguments, and then copy the result files 
out to the object store.

Code Structure
--------------

Main function of taskserver program (client interface and scheduler)
```
src/tes-server/
```

Main function of worker program
```
src/tes-worker/
```
 
Code related to the worker, include file mapper, local and swift file system clients, docker interfaces and worker thread manager
```
src/tes/worker/
```
 
The compiled copy of the Task Execution Schema protobuf
```
src/tes/ga4gh
```
 
BoltDB based implementation of the TES api as well as the scheduler API.
```
src/tes/server
```
 
The compiled copy of the scheduler API protobuf
```
src/tes/server/proto/
```
 
Python driven unit/conformance tests
```
tests
```
 
HTML and angular app for web view
```
share
```


Rebuilding Proto code
---------------------
First install protoc extentions (GRPC and GRPC gateway builder programs)
```
make depends
```
Rebuild auto-generated code
```
make proto_build
```


Running Tests
-------------
The integration testing is done using Python scripts to drive client API requests.
To run tests:
```
nosetests tests
```


