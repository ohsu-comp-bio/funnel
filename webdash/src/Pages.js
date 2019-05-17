import React from 'react';
import PropTypes from 'prop-types';
import Typography from '@material-ui/core/Typography';
import ReactJson from 'react-json-view';
import _ from "underscore";
import { TaskTable, NodeTable } from './Tables';

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
        return undefined;
      },
    );
};

class Task extends React.Component {
  state = {
    task: {},
  };

  componentDidMount() {
    var url = new URL("/v1/tasks/" + this.props.match.params.task_id, window.location.origin);
    get(url).then((task) => {
      //console.log("task:", task);
      this.setState({task: task});
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
        <Typography variant="h4" gutterBottom component="h2">    
          Task: {this.props.match.params.task_id}
        </Typography>
        <div style={{margin:"10px 0px"}}>{task}</div>
      </div>
    );
  };
};

class Node extends React.Component {
  state = {
    node: {},
  };

  componentDidMount() {
    var url = new URL("/v1/nodes/" + this.props.match.params.node_id, window.location.origin);
    get(url).then((node) => {
      //console.log("node:", node);
      this.setState({node: node});
    });
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
        <div style={{margin:"10px 0px"}}>{node}</div>
      </div>
    );
  };
};

class ServiceInfo extends React.Component {
  state = {
    info: {},
  };

  componentDidMount() {
    var url = new URL("/v1/tasks/service-info", window.location.origin);
    get(url).then((info) => {
      //console.log("service info:", info);
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
    return (
      <div>
        <Typography variant="h4" gutterBottom component="h2">    
          Service Info
        </Typography>
        <div style={{margin:"10px 0px"}}>{info}</div>
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
