import React from 'react';
import PropTypes from 'prop-types';
import Typography from '@material-ui/core/Typography';
import ReactJson from 'react-json-view';
import _ from "underscore";
import { NodeTable } from './NodeList';
import { SystemInfo } from './SystemInfo';
import { TaskTable } from './TaskList';
import { TaskInfo } from './TaskInfo';
import { renderCancelButton } from './utils';

class TaskList extends React.Component {
  constructor(props) {
    super(props)
    this.state = {
      pageSize: 50,
      pageToken: "",
      nextPageToken: "",
      prevPageToken: [],
      tasks: [],
    };
  };

  nextPage = () => {
    const current = this.state.pageToken;
    const next = this.state.nextPageToken;
    var prev = this.state.prevPageToken.concat(current);
    this.setState({
      prevPageToken: prev,
      pageToken: next,
    });
  };

  prevPage = () => {
    var prev = this.state.prevPageToken;
    const page = prev.pop();
    this.setState({
      pageToken: page,
      prevPageToken: prev,
    });
  };

  setPageSize = (event) => {
    this.setState({
      pageSize: event.target.value,
    });
  };

  listTasks() {
    var url = new URL("/v1/tasks" + window.location.search, window.location.origin);
    var params = url.searchParams;
    params.set("view", "BASIC");
    params.set("pageSize", this.state.pageSize);
    if (this.props.stateFilter !== "") {
      params.set("state", this.props.stateFilter);
    };
    for (var i = 0; i < this.props.tagsFilter.length; i++) {
      var tag = this.props.tagsFilter[i];
      if (tag.key !== "") {
        params.set("tags["+tag.key+"]", tag.value);
      };
    };
    if (this.state.pageToken !== "") {
      params.set("pageToken", this.state.pageToken);
    };
    //console.log("listTasks url:", url);
    fetch(url.toString())
      .then(response => response.json())
      .then(
        (result) => {
          //console.log("listTasks result:", result);
          var tasks = [];
          if (result.tasks !== undefined) {
            tasks = result.tasks;
          };
          var nextPageToken = "";
          if (result.nextPageToken !== undefined) {
            nextPageToken = result.nextPageToken;
          };
          this.setState({tasks: tasks, nextPageToken: nextPageToken});
        },
        (error) => {
          console.log("listTasks", url.toString(), "error:", error);
        },
      );
  };

  shouldComponentUpdate(nextProps, nextState) {
    if (!_.isEqual(this.props.stateFilter, nextProps.stateFilter)) {
      return true;
    };
    if (!_.isEqual(this.props.tagsFilter, nextProps.tagsFilter)) {
      return true;
    }
    if (!_.isEqual(this.state.pageToken, nextState.pageToken)) {
      return true;
    };
    if (!_.isEqual(this.state.pageSize, nextState.pageSize)) {
      return true;
    };
    if (!_.isEqual(this.state.tasks, nextState.tasks)) {
      return true;
    };
    return false;
  };

  render() {
    this.listTasks();
    //console.log("TaskList state:", this.state)
    //console.log("TaskList props:", this.props)
    return (
      <div>
        <Typography variant="h4" gutterBottom component="h2">
          Tasks
        </Typography>
        <TaskTable
         tasks={this.state.tasks}
         pageSize={this.state.pageSize}
         nextPageToken={this.state.nextPageToken}
         prevPageToken={this.state.prevPageToken}
         setPageSize={this.setPageSize}
         nextPage={this.nextPage}
         prevPage={this.prevPage}
        />
      </div>
    );
  };
};

TaskList.propTypes = {
  stateFilter: PropTypes.string.isRequired,
  tagsFilter: PropTypes.array.isRequired,
};

class NodeList extends React.Component {
  constructor(props) {
    super(props)
    this.state = {
      nodes: [],
    };
  };

  listNodes() {
    var url = new URL("/v1/nodes" + window.location.search, window.location.origin);
    //console.log("listNodes url:", url);
    fetch(url.toString())
      .then(response => response.json())
      .then(
        (result) => {
          //console.log("listNodes result:", result);
          if (result.nodes !== undefined) {
            this.setState({nodes: result.nodes});
          } else {
            this.setState({nodes: []});
          };
        },
        (error) => {
          console.log("listNodes", url.toString(), "error:", error);
        },
      );
  };

  shouldComponentUpdate(nextProps, nextState) {
    return false;
  };

  render() {
    this.listNodes();
    return (
      <div>
        <Typography variant="h4" gutterBottom component="h2">
          Nodes
        </Typography>
        <NodeTable nodes={this.state.nodes} />
      </div>
    );
  };
};

function get(url) {
  if (!url instanceof URL) {
    console.log("get error: expected URL object; got", url);
    return undefined;
  };
  var params = url.searchParams;
  params.set("view", "FULL");
  //console.log("get url:", url);
  return fetch(url.toString())
    .then(response => response.json())
    .then(
      (result) => {
        //console.log("get result:", result);
        return result;
      },
      (error) => {
        console.log("get", url.toString(), "error:", error);
        throw error
      },
    );
};

class Task extends React.Component {
  state = {
    //task: {},
    error: "",
    task: {
      "id": "bnj8hlnpbjg64189lu30",
      "state": "COMPLETE",
      "name": "sh -c 'echo starting; cat $file1 \u003e $file2; echo done'",
      "tags": {
        "tag-ONE": "TWO",
        "tag-THREE": "FOUR"
      },
      "volumes": ["/vol1", "/vol2"],
      "inputs": [
        {
          "name": "file1",
          "url": "file:///Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/README.md",
          "path": "/inputs/Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/README.md"
        }
      ],
      "outputs": [
        {
          "name": "stdout-0",
          "url": "file:///Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/test.stdout",
          "path": "/outputs/stdout-0"
        },
        {
          "name": "file2",
          "url": "file:///Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/test_out",
          "path": "/outputs/Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/test_out"
        }
      ],
      "resources": {
        "cpuCores": 2,
        "ramGb": 4,
        "diskGb": 10
      },
      "executors": [
        {
          "image": "ubuntu",
          "command": [
            "sh",
            "-c",
            "echo starting; cat $file1 \u003e $file2; echo done"
          ],
          "stdout": "/outputs/stdout-0",
          "env": {
            "file1": "/inputs/Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/README.md",
            "file2": "/outputs/Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/test_out"
          }
        }
      ],
      "logs": [
        {
          "logs": [
            {
              "startTime": "2019-12-03T08:09:58.524782-08:00",
              "endTime": "2019-12-03T08:10:04.209567-08:00",
              "stdout": "starting\ndone\n"
            }
          ],
          "metadata": {
            "hostname": "BICB230"
          },
          "startTime": "2019-12-03T08:09:58.516832-08:00",
          "endTime": "2019-12-03T08:10:04.216273-08:00",
          "outputs": [
            {
              "url": "file:///Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/test.stdout",
              "path": "/outputs/stdout-0",
              "sizeBytes": "14"
            },
            {
              "url": "file:///Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/test_out",
              "path": "/outputs/Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/test_out",
              "sizeBytes": "1209"
            }
          ],
          "systemLogs": [
            "level='info' msg='Version' timestamp='2019-12-03T08:09:58.5126-08:00' task_attempt='0' executor_index='0' GitCommit='a630947d' GitBranch='master' GitUpstream='git@github.com:ohsu-comp-bio/funnel.git' BuildDate='2019-01-29T00:50:30Z' Version='0.9.0'",
            "level='info' msg='download started' timestamp='2019-12-03T08:09:58.521417-08:00' task_attempt='0' executor_index='0' url='file:///Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/README.md'",
            "level='info' msg='download finished' timestamp='2019-12-03T08:09:58.523199-08:00' task_attempt='0' executor_index='0' url='file:///Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/README.md' size='1209' etag=''",
            "level='info' msg='Running command' timestamp='2019-12-03T08:10:02.83215-08:00' task_attempt='0' executor_index='0' cmd='docker run -i --read-only --rm -e file1=/inputs/Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/README.md -e file2=/outputs/Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/test_out --name bnj8hlnpbjg64189lu30-0 -v /Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/funnel-work-dir/bnj8hlnpbjg64189lu30/tmp:/tmp:rw -v /Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/funnel-work-dir/bnj8hlnpbjg64189lu30/inputs/Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/README.md:/inputs/Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/README.md:ro -v /Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/funnel-work-dir/bnj8hlnpbjg64189lu30/outputs:/outputs:rw ubuntu sh -c echo starting; cat $file1 \u003e $file2; echo done'",
            "level='info' msg='upload started' timestamp='2019-12-03T08:10:04.211287-08:00' task_attempt='0' executor_index='0' url='file:///Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/test_out'",
            "level='info' msg='upload finished' timestamp='2019-12-03T08:10:04.213672-08:00' task_attempt='0' executor_index='0' size='14' url='file:///Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/test.stdout' etag=''"
          ]
        }
      ],
      "creationTime": "2019-12-03T08:09:58.506338-08:00"
    }
  };

  componentDidMount() {
    var url = new URL("/v1/tasks/" + this.props.match.params.task_id, window.location.origin);
    get(url).then(
      (task) => {
      //console.log("task:", task);
      this.setState({task: task});
    },
      (error) => {
        this.setState({error: "Error: " + error.toString()});
      });
  };

  render() {
    var task = (
      <ReactJson
       src={this.state.task}
       theme={"rjv-default"}
       name={false}
       displayObjectSize={false}
       displayDataTypes={false}
       enableClipboard={true}
       collapsed={false}
      />
    );
    return (
      <div>
        <Typography variant="h4" gutterBottom>
          Task: {this.props.match.params.task_id}
        </Typography>
        <Typography variant="h5" gutterBottom color="textSecondary">
          {this.state.error}
        </Typography>
        {renderCancelButton(this.state.task)}
        {/* <div style={{margin:"10px 0px"}}>{task}</div> */}
        <TaskInfo task={this.state.task} />
      </div>
    );
  };
};

class Node extends React.Component {
  state = {
    node: {},
    error: ""
  };

  componentDidMount() {
    var url = new URL("/v1/nodes/" + this.props.match.params.node_id, window.location.origin);
    get(url).then(
      (node) => {
        //console.log("node:", node);
        this.setState({node: node});
      },
      (error) => {
        this.setState({error: "Error: " + error.toString()});
      });;
  };

  render() {
    var node = (
      <ReactJson
       src={this.state.node}
       theme={"rjv-default"}
       name={false}
       displayObjectSize={false}
       displayDataTypes={false}
       enableClipboard={true}
       collapsed={false}
      />
    );
    return (
      <div>
        <Typography variant="h4" gutterBottom component="h2">
          Node: {this.props.match.params.node_id}
        </Typography>
        <Typography variant="h5" gutterBottom color="textSecondary">
          {this.state.error}
        </Typography>
        <div style={{margin:"10px 0px"}}>{node}</div>
      </div>
    );
  };
};

class ServiceInfo extends React.Component {
  state = {
    //info: {},
    error: "",
    info: {
      "name": "Funnel",
      "doc": "git commit: a630947d\ngit branch: master\ngit upstream: git@github.com:ohsu-comp-bio/funnel.git\nbuild date: 2019-01-29T00:50:30Z\nversion: 0.9.0",
      "taskStateCounts": {
        "CANCELED": 0,
        "COMPLETE": 0,
        "EXECUTOR_ERROR": 0,
        "INITIALIZING": 0,
        "PAUSED": 0,
        "QUEUED": 0,
        "RUNNING": 0,
        "SYSTEM_ERROR": 0,
        "UNKNOWN": 0
      }
    }
  };

  componentDidMount() {
    var url = new URL("/v1/tasks/service-info", window.location.origin);
    get(url).then(
      (info) => {
      //console.log("service info:", info);
      this.setState({info: info});
      },
      (error) => {
        this.setState({error: "Error: " + error.toString()});
      });
  }

  render() {
    return (
      <div>
        <Typography variant="h4" gutterBottom component="h2">
          Service Info
        </Typography>
        <SystemInfo info={this.state.info} />
      </div>
    );
  };
};

function NoMatch() {
  return (
    <Typography variant="h4" gutterBottom component="h2" style={{color: "red"}}>
      404 Not Found
    </Typography>
 );
};

export {TaskList, Task, NodeList, Node, ServiceInfo, NoMatch};
