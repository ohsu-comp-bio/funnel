package server

import (
	"io"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/ohsu-comp-bio/funnel/tes"
	"google.golang.org/protobuf/encoding/protojson"
)

// MarshalNew is a shim class to 'fix' outgoing streamed messages
// in the default implementation, grpc-gateway wraps the individual messages
// of the stream with a {"result" : <value>}. The cleaner idendifies that and
// removes the wrapper
type MarshalNew struct {
	m runtime.Marshaler
}

func NewMarshaler() runtime.Marshaler {
	return &MarshalNew{
		m: &runtime.JSONPb{
			protojson.MarshalOptions{
				Indent:          "  ",
				EmitUnpopulated: true,
				UseProtoNames:   true,
			},
			protojson.UnmarshalOptions{},
		},
	}
}

// ContentType return content type of marshler
func (mclean *MarshalNew) ContentType(i interface{}) string {
	return mclean.m.ContentType(i)
}

// Marshal serializes v into a JSON encoded byte array. If v is of
// type `proto.Message` the then field "result" is extracted and returned by
// itself. This is mainly to get around a weird behavior of the GRPC gateway
// streaming output
func (mclean *MarshalNew) Marshal(v interface{}) ([]byte, error) {
    // Type assertion to get the underlying *tes.Task
    task, ok := v.(*tes.Task)
    if !ok {
        // v is not of type *tes.Task
		return mclean.m.Marshal(v)
    }
	if task.CreationTime == "" {
		// view = "MINIMAL"
		min := &tes.TaskMin{
			Id:    task.Id,
			State: task.State,
		}
		return mclean.m.Marshal(min)
	} else if len(task.Logs[0].SystemLogs) == 0 {
		// view = "BASIC"
		executors := []*tes.ExecutorBasic{}
		for _, executor := range task.Executors {
			executors = append(executors, &tes.ExecutorBasic{
				Command: 	 executor.Command,
				Env:    	 executor.Env,
				IgnoreError: executor.IgnoreError,
				Image: 		 executor.Image,
				Stdin: 		 executor.Stdin,
				Workdir: 	 executor.Workdir,
			})
		}

		inputs := []*tes.InputBasic{}
		for _, input := range task.Inputs {
			inputs = append(inputs, &tes.InputBasic{
				Description: input.Description,
				Name: 		 input.Name,
				Path: 		 input.Path,
				Streamable:  input.Streamable,
				Type: 		 input.Type,
				Url: 		 input.Url,
			})
		}

		logs := []*tes.TaskLogBasic{}
		for _, log := range task.Logs {
			logs = append(logs, &tes.TaskLogBasic{
				EndTime:   log.EndTime,
				Logs:      log.Logs,
				Metadata:  log.Metadata,
				Outputs:   log.Outputs,
				StartTime: log.StartTime,
			})
		}

		basic := &tes.TaskBasic{
			CreationTime: task.CreationTime,
			Description:  task.Description,
			Executors:    executors,
			Id:			  task.Id,
			Inputs:		  inputs,
			Logs:	 	  logs,
			Name: 		  task.Name,
			Outputs:	  task.Outputs,
			Resources: 	  task.Resources,
			State: 		  task.State,
			Tags: 		  task.Tags,
			Volumes: 	  task.Volumes,
		}

		return mclean.m.Marshal(basic)
	} else {
		// view = "FULL"
		return mclean.m.Marshal(v)
	}
}

// NewDecoder shims runtime.Marshaler.NewDecoder
func (mclean *MarshalNew) NewDecoder(r io.Reader) runtime.Decoder {
	return mclean.m.NewDecoder(r)
}

// NewEncoder shims runtime.Marshaler.NewEncoder
func (mclean *MarshalNew) NewEncoder(w io.Writer) runtime.Encoder {
	return mclean.m.NewEncoder(w)
}

// Unmarshal shims runtime.Marshaler.Unmarshal
func (mclean *MarshalNew) Unmarshal(data []byte, v interface{}) error {
	return mclean.m.Unmarshal(data, v)
}
