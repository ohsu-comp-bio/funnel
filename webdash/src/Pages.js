import React from 'react';
import PropTypes from 'prop-types';
import Typography from '@material-ui/core/Typography';
import { TaskTable, SimpleTable } from './Tables';

class TaskList extends React.Component {
  constructor(props) {
    super(props)
    this.state = {
      tasks: [],
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
          }
        },
        (error) => {
          console.log("listTasks error:", error)
        }
      )
  };

  componentDidMount() {
    this.listTasks();
  }

  shouldComponentUpdate(nextProps, nextState) {
    if (this.props.stateFilter !== nextProps.stateFilter) {
      return true
    }
    // TODO: return true on tags filter change
    if (this.state.tasks.length === undefined || this.state.tasks.length === 0) {
      return true
    }
    if (this.state.tasks.length && nextState.tasks.length) {
      // if the first task is different we should update the table
      return this.state.tasks[0].id !== nextState.tasks[0].id
    }
    return false
  }

  render() {
    console.log("TaskList props:", this.props)
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
          }
        },
        (error) => {
          console.log("listNodes error:", error)
        }
      )
  };

  componentDidMount() {
    this.listNodes();
  }

  shouldComponentUpdate(nextProps, nextState) {
    if (this.state.nodes.length === undefined || this.state.nodes.length === 0) {
      return true
    }
    if (this.state.nodes.length && nextState.nodes.length) {
      // if the first node is different we should update the table
      return this.state.nodes[0].id !== nextState.nodes[0].id
    }
    return false
  }

  render() {
    return (
      <div>
        <Typography variant="h4" gutterBottom component="h2">
          Nodes
        </Typography>
        <SimpleTable />
      </div>
    );
  }
}

function get(url) {
  if (typeof url !== Object) {
    console.log("get error: expected URL object; got", url)
    return undefined
  }
  var params = url.searchParams
  params.set("view", "FULL")
  console.log("get url:", url)
  fetch(url.toString())
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

function Task({ match }) {
  var url = new URL("/v1/tasks/" + match.params.task_id, window.location.origin);
  var task = get(url);
  if (task === undefined) {
    task = NoMatch();
  }
  return (
    <div>
      <Typography variant="h4" gutterBottom component="h2">    
        Task: {match.params.task_id}
      </Typography>
      <div style={{margin:"10px 0px"}}>{task}</div>
    </div>
  )
}

function Node({ match }) {
  var url = new URL("/v1/nodes/" + match.params.node_id, window.location.origin);
  var node = get(url);
  if (node === undefined) {
    node = NoMatch();
  }
  return (
    <div>
      <Typography variant="h4" gutterBottom component="h2">    
        Node: {match.params.node_id}
      </Typography>
      <div style={{margin:"10px 0px"}}>{node}</div>
    </div>
  )
}

function ServiceInfo() {
  var url = new URL("/v1/tasks/service-info", window.location.origin);
  var info = get(url);
  if (info === undefined) {
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

function NoMatch() {
  return (
    <Typography variant="h4" gutterBottom component="h2" style={{color: "red"}}>
      404 Not Found
    </Typography>
 )
}

export {TaskList, Task, NodeList, Node, ServiceInfo, NoMatch};
