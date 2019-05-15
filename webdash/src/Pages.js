import React from 'react';
import PropTypes from 'prop-types';
import Typography from '@material-ui/core/Typography';
import ReactJson from 'react-json-view';
import _ from "underscore";
import { TaskTable, NodeTable, SimpleTable } from './Tables';

class TaskList extends React.Component {
  constructor(props) {
    super(props)
    this.state = {
      tasks: [],
      // tasks: [
      //   {
      //     id: "bij587vpbjg2fb6m0beg",
      //     state: "CANCELED",
      //     name: "sh -c 'echo 1 && sleep 240'",
      //     executors:[{image: "alpine",
      //                 command: ["sh","-c","echo 1 && sleep 240"]}],
      //     logs: [{logs: [{startTime: "2019-04-04T11:59:43.977367-07:00",
      //                     stdout: "1\n"}],
      //             metadata:{hostname: "BICB230"},
      //             startTime: "2019-04-04T11:59:43.854152-07:00"}],
      //     creationTime: "2019-04-04T11:59:43.793262-07:00"
      //   }
      // ]
    };
  };

  listTasks = () => {
    var url = new URL("/v1/tasks" + window.location.search, window.location.origin);
    var params = url.searchParams
    params.set("view", "BASIC")
    params.set("pageSize", 50)
    if (this.props.stateFilter !== undefined && this.props.stateFilter !== "") {
      params.set("state", this.props.stateFilter)
    }
    console.log("listTasks url:", url)
    fetch(url.toString())
      .then(response => response.json())
      .then(
        (result) => {
          console.log("listTasks result:", result)
          if (result.tasks !== undefined) {
            this.setState({tasks: result.tasks})
          } else {
            this.setState({tasks: []})
          }
        },
        (error) => {
          console.log("listTasks error:", error)
        }
      )
  };

  shouldComponentUpdate(nextProps, nextState) {
    if (!_.isEqual(this.props.stateFilter, nextProps.stateFilter)) {
      return true
    }
    if (!_.isEqual(this.props.tagsFilter, nextProps.tagsFilter)) {
      return true
    }
    if (!_.isEqual(this.state.tasks, nextState.tasks)) {
      return true
    }
    return false
  }

  render() {
    this.listTasks();
    return (
      <div>
        <Typography variant="h4" gutterBottom component="h2">
          Tasks
        </Typography>
        <TaskTable tasks={this.state.tasks}/>
      </div>
    )
  }
}

TaskList.propTypes = {
  stateFilter: PropTypes.string.isRequired,
  tagsFilter: PropTypes.object.isRequired,
};

class NodeList extends React.Component {
  constructor(props) {
    super(props)
    this.state = {
      nodes: [],
    };
  };
  
  listNodes = () => {
    var url = new URL("/v1/nodes" + window.location.search, window.location.origin);
    console.log("listNodes url:", url)
    fetch(url.toString())
      .then(response => response.json())
      .then(
        (result) => {
          console.log("listNodes result:", result)
          if (result.nodes !== undefined) {
            this.setState({nodes: result.nodes})
          } else {
            this.setState({nodes: []})
          }
        },
        (error) => {
          console.log("listNodes error:", error)
        }
      )
  };
  
  shouldComponentUpdate(nextProps, nextState) {
    if (!_.isEqual(this.state.nodes, nextState.nodes)) {
      return true
    }
    return false
  }

  render() {
    this.listNodes();
    return (
      <div>
        <Typography variant="h4" gutterBottom component="h2">
          Nodes
        </Typography>
        <NodeTable nodes={this.state.nodes} />
        <br/>
        <SimpleTable />
      </div>
    );
  }
}

function get(url) {
  if (!url instanceof URL) {
    console.log("get error: expected URL object; got", url)
    return undefined
  }
  var params = url.searchParams
  params.set("view", "FULL")
  console.log("get url:", url)
  return fetch(url.toString())
    .then(response => response.json())
    .then(
      (result) => {
        console.log("get result:", result)
        return result
      },
      (error) => {
        console.log("get error:", error)
        return undefined
      }
    )
}

class Task extends React.Component {
  state = {
    task: undefined,
    // task: {
    //   id: "bij587vpbjg2fb6m0beg",
    //   state: "CANCELED",
    //   name: "sh -c 'echo 1 && sleep 240'",
    //   executors:[{image: "alpine",
    //               command: ["sh","-c","echo 1 && sleep 240"]}],
    //   logs: [{logs: [{startTime: "2019-04-04T11:59:43.977367-07:00",
    //                   stdout: "1\n"}],
    //           metadata:{hostname: "BICB230"},
    //           startTime: "2019-04-04T11:59:43.854152-07:00"}],
    //   creationTime: "2019-04-04T11:59:43.793262-07:00"
    // }
  };

  componentDidMount() {
    var url = new URL("/v1/tasks/" + this.props.match.params.task_id, window.location.origin);
    get(url).then((task) => {
      console.log("task:", task)
      this.setState({task: task});
    });
  }
    
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
    if (this.state.task === undefined) {
      task = NoMatch();
    }
    return (
      <div>
        <Typography variant="h4" gutterBottom component="h2">    
          Task: {this.props.match.params.task_id}
        </Typography>
        <div style={{margin:"10px 0px"}}>{task}</div>
      </div>
    );
  }
}

class Node extends React.Component {
  state = {
    node: undefined,
  };

  componentDidMount() {
    var url = new URL("/v1/nodes/" + this.props.match.params.node_id, window.location.origin);
    get(url).then((node) => {
      console.log("node:", node)
      this.setState({node: node});
    });
  }
    
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
    if (this.state.node === undefined) {
      node = NoMatch();
    }
    return (
      <div>
        <Typography variant="h4" gutterBottom component="h2">    
          Node: {this.props.match.params.node_id}
        </Typography>
        <div style={{margin:"10px 0px"}}>{node}</div>
      </div>
    )
  }
}

class ServiceInfo extends React.Component {
  state = {
    info: undefined,
  };

  componentDidMount() {
    var url = new URL("/v1/tasks/service-info", window.location.origin);
    get(url).then((info) => {
      console.log("service info:", info)
      this.setState({info: info});
    });
  }
 
  render() {
    var info = (
      <ReactJson 
       src={this.state.info}
       theme={"rjv-default"}
       name={false}
       displayObjectSize={false}
       displayDataTypes={false}
       enableClipboard={true}
       collapsed={false}       
      />
    );
    if (this.state.info === undefined) {
      info = NoMatch();
    }
    return (
      <div>
        <Typography variant="h4" gutterBottom component="h2">    
          Service Info
        </Typography>
        <div style={{margin:"10px 0px"}}>{info}</div>
      </div>
    )
  }
}

function NoMatch() {
  return (
    <Typography variant="h4" gutterBottom component="h2" style={{color: "red"}}>
      404 Not Found
    </Typography>
 )
}

export {TaskList, Task, NodeList, Node, ServiceInfo, NoMatch};
