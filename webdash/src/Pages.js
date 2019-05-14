import React from 'react';
import Typography from '@material-ui/core/Typography';
import TaskTable from './Tables';

class TaskList extends React.Component {
  constructor(props) {
    super(props)
    this.state = {
      tasks: [],
    };
  };

  getTasks = () => {
    var url = new URL(window.location.pathname, window.location.origin);
    var params = url.searchParams
    params.set("view", "BASIC")
    params.set("pageSize", 50)
    console.log("getTasks url:", url)
    fetch(url.toString())
      .then(response => response.json())
      .then(
        (result) => {
          console.log("getTasks result:", result)
          this.setState({tasks: result.tasks})
        },
        (error) => {
          console.log("getTasks error:", error)
        }
      )
  };

  render() {
    console.log("TaskList props:", this.props)
    this.getTasks();
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

function Task({ match }) {
  return (
    <Typography variant="h4" gutterBottom component="h2">    
      Task: {match.params.task_id}
    </Typography>
 )
}

function NodeList() {
  return (
    <div>
      <Typography variant="h4" gutterBottom component="h2">
        Nodes
      </Typography>
      <TaskTable tasks={[]} />
    </div>
 )
}

function Node({ match }) {
  return (
    <Typography variant="h4" gutterBottom component="h2">    
      Node: {match.params.task_id}
    </Typography>
 )
}

function ServiceInfo() {
  return (
    <Typography variant="h4" gutterBottom component="h2">
      Service Info
    </Typography>
 )
}

function NoMatch() {
  return (
    <Typography variant="h4" gutterBottom component="h2">
      404 Not Found
    </Typography>
 )
}

export {TaskList, Task, NodeList, Node, ServiceInfo, NoMatch};
