import React from 'react';
import Typography from '@material-ui/core/Typography';
import SimpleTable from './SimpleTable';

export function TaskList() {
  return (
    <div>
      <Typography variant="h4" gutterBottom component="h2">
        Tasks
      </Typography>
      <SimpleTable />
    </div>
  )
}

export function Task({ match }) {
  return (
    <Typography variant="h4" gutterBottom component="h2">    
      Task: {match.params.task_id}
    </Typography>
 )
}

export function NodeList() {
  return (
    <div>
      <Typography variant="h4" gutterBottom component="h2">
        Nodes
      </Typography>
      <SimpleTable />
    </div>
 )
}

export function Node({ match }) {
  return (
    <Typography variant="h4" gutterBottom component="h2">    
      Node: {match.params.task_id}
    </Typography>
 )
}

export function ServiceInfo() {
  return (
    <Typography variant="h4" gutterBottom component="h2">
      Service Info
    </Typography>
 )
}

export function NoMatch() {
  return (
    <Typography variant="h4" gutterBottom component="h2">
      404 Not Found
    </Typography>
 )
}
