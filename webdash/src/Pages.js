import React from 'react';
import Typography from '@material-ui/core/Typography';
import SimpleTable from './SimpleTable';

class TaskList extends React.Component {

  render() {
    console.log("TaskList props", this.props)
    return (
      <div>
        <Typography variant="h4" gutterBottom component="h2">
          Tasks
        </Typography>
        <SimpleTable />
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
      <SimpleTable />
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
