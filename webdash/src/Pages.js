import React from 'react';
import Typography from '@material-ui/core/Typography';

export function TaskList() {
  return (    
    <Typography variant="h4" gutterBottom component="h2">
      Tasks
    </Typography>
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
    <Typography variant="h4" gutterBottom component="h2">
      Nodes
    </Typography>
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
      ServiceInfo
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
