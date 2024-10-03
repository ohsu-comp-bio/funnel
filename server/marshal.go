package server

import (
	"io"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/ohsu-comp-bio/funnel/tes"
	"google.golang.org/protobuf/encoding/protojson"
)

// CustomMarshal is a custom marshaler for the GRPC gateway that returns the required fields based on the View value:
// - View_MINIMAL: returns only the id and state fields
// - View_BASIC: returns the id, state, creation_time, description, name, inputs, outputs, resources, tags, volumes, executors, and logs fields
// - View_FULL: returns all fields
//
// This could be improved by updating the generated protobuf code to include the View field in the Task struct
// Related discussion: https://github.com/ohsu-comp-bio/funnel/pull/716#discussion_r1375155983
type CustomMarshal struct {
	m runtime.Marshaler
}

func NewMarshaler() runtime.Marshaler {
	return &CustomMarshal{
		m: &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				Indent:          "  ",
				EmitUnpopulated: true,
				UseProtoNames:   true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{},
		},
	}
}

// ContentType return content type of marshler
func (marshal *CustomMarshal) ContentType(i interface{}) string {
	return marshal.m.ContentType(i)
}

// Marshal serializes v into a JSON encoded byte array. If v is of
// type `proto.Message` the then field "result" is extracted and returned by
// itself. This is mainly to get around a weird behavior of the GRPC gateway
// streaming output
func (mclean *CustomMarshal) Marshal(v interface{}) ([]byte, error) {

	list, ok := v.(*tes.ListTasksResponse)
	if ok {
		// v is of type *tes.ListTasksResponse
		return mclean.MarshalList(list)
	}

	task, ok := v.(*tes.Task)
	if ok {
		// v is of type *tes.Task
		return mclean.MarshalTask(task)
	}

	return mclean.m.Marshal(v)
}

func (mclean *CustomMarshal) MarshalTask(task *tes.Task) ([]byte, error) {
	view, _ := mclean.DetectView(task)
	newTask := mclean.TranslateTask(task, view)
	return mclean.m.Marshal(newTask)
}

func (mclean *CustomMarshal) MarshalList(list *tes.ListTasksResponse) ([]byte, error) {
	if len(list.Tasks) == 0 {
		return mclean.m.Marshal(list)
	}

	task := list.Tasks[0]
	view, _ := mclean.DetectView(task)

	if view == tes.View_MINIMAL {
		minList := &tes.ListTasksResponseMin{}
		for _, task := range list.Tasks {
			minTask := mclean.TranslateTask(task, view).(*tes.TaskMin)
			minList.Tasks = append(minList.Tasks, minTask)
		}
		return mclean.m.Marshal(minList)
	}

	if view == tes.View_BASIC {
		basicList := &tes.ListTasksResponseBasic{}
		for _, task := range list.Tasks {
			basicTask := mclean.TranslateTask(task, view).(*tes.TaskBasic)
			basicList.Tasks = append(basicList.Tasks, basicTask)
		}
		return mclean.m.Marshal(basicList)
	}

	return mclean.m.Marshal(list)
}

func (mclean *CustomMarshal) DetectView(task *tes.Task) (tes.View, error) {
	if task.CreationTime == "" {
		// return a MINIMAL view
		return tes.View_MINIMAL, nil
	}

	if len(task.Logs[0].SystemLogs) == 0 {
		return tes.View_BASIC, nil
	}

	// view = "FULL"
	return tes.View_FULL, nil
}

func (mclean *CustomMarshal) TranslateTask(task *tes.Task, view tes.View) interface{} {
	// view = "MINIMAL"
	if view == tes.View_MINIMAL {
		min := &tes.TaskMin{
			Id:    task.Id,
			State: task.State,
		}
		return min
	}

	// view = "BASIC"
	if view == tes.View_BASIC {
		executors := []*tes.ExecutorBasic{}
		for _, executor := range task.Executors {
			executors = append(executors, &tes.ExecutorBasic{
				Command:     executor.Command,
				Env:         executor.Env,
				IgnoreError: executor.IgnoreError,
				Image:       executor.Image,
				Stdin:       executor.Stdin,
				Workdir:     executor.Workdir,
			})
		}

		inputs := []*tes.InputBasic{}
		for _, input := range task.Inputs {
			inputs = append(inputs, &tes.InputBasic{
				Description: input.Description,
				Name:        input.Name,
				Path:        input.Path,
				Streamable:  input.Streamable,
				Type:        input.Type,
				Url:         input.Url,
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
			Id:           task.Id,
			Inputs:       inputs,
			Logs:         logs,
			Name:         task.Name,
			Outputs:      task.Outputs,
			Resources:    task.Resources,
			State:        task.State,
			Tags:         task.Tags,
			Volumes:      task.Volumes,
		}

		return basic
	}

	// view = "FULL"
	return task
}

// NewDecoder shims runtime.Marshaler.NewDecoder
func (mclean *CustomMarshal) NewDecoder(r io.Reader) runtime.Decoder {
	return mclean.m.NewDecoder(r)
}

// NewEncoder shims runtime.Marshaler.NewEncoder
func (mclean *CustomMarshal) NewEncoder(w io.Writer) runtime.Encoder {
	return mclean.m.NewEncoder(w)
}

// Unmarshal shims runtime.Marshaler.Unmarshal
func (mclean *CustomMarshal) Unmarshal(data []byte, v interface{}) error {
	return mclean.m.Unmarshal(data, v)
}
