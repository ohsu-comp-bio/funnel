import React from 'react';
import './App.css';
import { BrowserRouter as Router, Route, Switch } from "react-router-dom";

function App() {
  return (
      <Router>
        <div>
          <SideBar />
          <Switch>
            <Route exact path="/v1/tasks" component={TaskList} />
            <Route exact path="/" component={TaskList} />
            <Route exact path="/tasks" component={TaskList} />
            <Route exact path="/v1/tasks/:task_id" component={Task} />
            <Route exact path="/tasks/:task_id" component={Task} />
            <Route exact path="/v1/nodes" component={NodeList} />
            <Route exact path="/nodes" component={NodeList} />
            <Route exact path="/v1/nodes/:node_id" component={Node} />
            <Route exact path="/nodes/:node_id" component={Node} />
            <Route exact path="/v1/tasks/service-info" component={ServiceInfo} />
            <Route exact path="/tasks/service-info" component={ServiceInfo} />
            <Route exact path="/service-info" component={ServiceInfo} />
            <Route component={NoMatch} />
          </Switch>
        </div>
      </Router>
  );
}

function SideBar() {
  return (
    <el>SideBar</el>
 )
}

function TaskList() {
  return (
    <h2>Tasks</h2>
 )
}

function Task({ match }) {
  return (
    <h2>Task: {match.params.task_id}</h2>
 )
}

function NodeList() {
  return (
    <h2>Nodes</h2>
 )
}

function Node({ match }) {
  return (
    <h2>Node: {match.params.node_id}</h2>
 )
}

function ServiceInfo() {
  return (
    <h2>Service Info</h2>
 )
}

function NoMatch() {
  return (
    <h2>404 Not Found</h2>
 )
}


export default App;
