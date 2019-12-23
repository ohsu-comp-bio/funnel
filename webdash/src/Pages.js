import React from 'react';
import PropTypes from 'prop-types';
import Typography from '@material-ui/core/Typography';
import ReactJson from 'react-json-view';
import { useParams} from "react-router";
import _ from "underscore";
import { NodeTable } from './NodeList';
import { NodeInfo } from './NodeInfo';
import { SystemInfo } from './SystemInfo';
import { TaskTable } from './TaskList';
import { TaskInfo } from './TaskInfo';
import { get, renderCancelButton } from './utils';
import { SimpleTabs } from './Tabs';
//import { example_task, example_node, example_service_info, example_task_list, example_node_list } from './ExampleData.js';

class TaskList extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      pageSize: 25,
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

function NodeList() {
  const [nodes, setNodes] = React.useState([]);
  //const [nodes, setNodes] = React.useState(example_node_list);

  React.useEffect(() => {
    var url = new URL("/v1/nodes", window.location.origin);
    // console.log("listNodes url:", url);
    fetch(url.toString())
      .then(response => response.json())
      .then(
        (result) => {
          if (result.nodes !== undefined) {
            setNodes(result.nodes);
          }
        },
        (error) => {
          console.log("listNodes err:", error.toString());
        },
      );
  });

  return (
    <div>
      <Typography variant="h4" gutterBottom component="h2">
        Nodes
      </Typography>
      <NodeTable nodes={nodes} />
    </div>
  );
};

function Task() {
  let { task_id } = useParams();
  const [task, setTask] = React.useState({});
  //const [task, setTask] = React.useState(example_task);
 
  React.useEffect(() => {
    var url = new URL("/v1/tasks/" + task_id, window.location.origin);
    get(url).then(
      (task) => {
      setTask(task);
    });
  }, [task_id]);

  const json = (
    <ReactJson
     src={task}
     theme={"rjv-default"}
     name={false}
     displayObjectSize={false}
     displayDataTypes={false}
     enableClipboard={true}
     collapsed={false}
    />
  );

  const header = (
    <div>
      <Typography variant="h4" gutterBottom>
        Task: {task_id}
      </Typography>
      {renderCancelButton(task)}
    </div>
  );

  return (
    SimpleTabs(header, <TaskInfo task={task} />, json)
  );
};

function Node() {
  let { node_id } = useParams();
  const [node, setNode] = React.useState({});
  //const [node, setNode] = React.useState(example_node);

  React.useEffect(() => {
    var url = new URL("/v1/nodes/" + node_id, window.location.origin);
    get(url).then(
      (node) => {
      setNode(node);
    });
  }, [node_id]);

  const json = (
    <ReactJson
      src={node}
      theme={"rjv-default"}
      name={false}
      displayObjectSize={false}
      displayDataTypes={false}
      enableClipboard={true}
      collapsed={false}
    />
  );

  const header = (
    <div>
      <Typography variant="h4" gutterBottom component="h2">
        Node: {node_id}
      </Typography>
    </div>
  );

  return (
    SimpleTabs(header, <NodeInfo node={node} />, json)
  );
};

function ServiceInfo() {
  const [info, setInfo] = React.useState({});
  //const [info, setInfo] = React.useState(example_service_info);

  React.useEffect(() => {
    var url = new URL("/v1/tasks/service-info", window.location.origin);
    get(url).then(
      (info) => {
      setInfo(info);
      });
  });
  
  const json = (
    <ReactJson
      src={info}
      theme={"rjv-default"}
      name={false}
      displayObjectSize={false}
      displayDataTypes={false}
      enableClipboard={true}
      collapsed={false}
    />
  );

  const header = (
    <div>
      <Typography variant="h4" gutterBottom component="h2">
        System Info
      </Typography>
    </div>
  );

  return (
    SimpleTabs(header, <SystemInfo info={info} />, json)
  );
};

function NoMatch() {
  return (
    <Typography variant="h4" gutterBottom component="h2" style={{color: "red"}}>
      404 Not Found
    </Typography>
 );
};

export {TaskList, Task, NodeList, Node, ServiceInfo, NoMatch};
