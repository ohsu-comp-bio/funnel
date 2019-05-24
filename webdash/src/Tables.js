import React from 'react';
import PropTypes from 'prop-types';
import { withStyles } from '@material-ui/core/styles';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';
import TableFooter from '@material-ui/core/TableFooter';
import TablePagination from '@material-ui/core/TablePagination';
import Paper from '@material-ui/core/Paper';
import IconButton from '@material-ui/core/IconButton';
import KeyboardArrowLeft from '@material-ui/icons/KeyboardArrowLeft';
import KeyboardArrowRight from '@material-ui/icons/KeyboardArrowRight';

const styles = {
  root: {
    width: '100%',
    overflowX: 'auto',
  },
  table: {
    minWidth: 700,
  },
};

class NodeTableRaw extends React.Component {
  nTasks(node) {
    if (node.task_ids !== undefined) {
      return node.task_ids.length
    }
    return 0
  }

  render() {
    const { classes } = this.props;
    return (
      <Paper className={classes.root}>
        <Table className={classes.table}>
          <TableHead>
            <TableRow>
              <TableCell>ID</TableCell>
              <TableCell>State</TableCell>
              <TableCell>Tasks</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {this.props.nodes.map(n => (
              <TableRow key={n.id}>
                <TableCell><a href={"/nodes/" + n.id}>{ n.hostname || n.id}</a></TableCell>
                <TableCell>{n.state}</TableCell>
                <TableCell>{this.nTasks(n)}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Paper>
    );
  }
}

NodeTableRaw.propTypes = {
  classes: PropTypes.object.isRequired,
  nodes: PropTypes.array.isRequired,
};

class TaskTableRaw extends React.Component {

  formatElapsedTime(miliseconds) {
    var days, hours, minutes, seconds, total_hours, total_minutes, total_seconds;

    total_seconds = parseInt(Math.floor(miliseconds / 1000));
    total_minutes = parseInt(Math.floor(total_seconds / 60));
    total_hours = parseInt(Math.floor(total_minutes / 60));

    seconds = parseInt(total_seconds % 60);
    minutes = parseInt(total_minutes % 60);
    hours = parseInt(total_hours % 24);
    days = parseInt(Math.floor(total_hours / 24));

    var time = "";
    if (days > 0) {
      time += days + "d "
    }
    if (hours > 0 || days > 0) {
      time += hours + "h "
    }
    if (minutes > 0 || hours > 0) {
      time += minutes + "m "
    }
    if (seconds > 0 || minutes > 0) {
      time += seconds + "s"
    }
    if (time === "") {
      time = "< 1s";
    }
    return time;
  }

  elapsedTime(task) {
    if (task.logs && task.logs.length) {
      if (task.logs[0].startTime) {
        var now = new Date();
        if (this.isDone(task)) {
          if (task.logs[0].endTime) {
            now = Date.parse(task.logs[0].endTime);
          } else {
            return "--";
          }
        }
        var started = Date.parse(task.logs[0].startTime);
        var elapsed = now - started;
        return this.formatElapsedTime(elapsed);
      }
    }
    return "--";
  }

  creationTime(task) {
    if (task.creationTime) {
      var created = new Date(task.creationTime);
      var options = {
        weekday: 'short',  month: 'short', day: 'numeric',
        hour: 'numeric', minute: 'numeric'
      };
      return created.toLocaleDateString("en-US", options);
    }
    return "--";
  }

  isDone(task) {
    return task.state === "COMPLETE" || task.state === "EXECUTOR_ERROR" || task.state === "CANCELED" || task.state === "SYSTEM_ERROR";
  }

  render() {
    //console.log("TaskTable props:", this.props)
    const { classes } = this.props;
    return (
      <Paper className={classes.root}>
        <Table className={classes.table}>
          <TableHead>
            <TableRow>
              <TableCell>ID</TableCell>
              <TableCell>State</TableCell>
              <TableCell>Name</TableCell>
              <TableCell>Created At</TableCell>
              <TableCell>Elapsed Time</TableCell>
              <TableCell></TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {this.props.tasks.map(t => (
              <TableRow key={t.id}>
                <TableCell><a href={"/tasks/" + t.id}>{t.id}</a></TableCell>
                <TableCell>{t.state}</TableCell>
                <TableCell>{t.name}</TableCell>
                <TableCell>{this.creationTime(t)}</TableCell>
                <TableCell>{this.elapsedTime(t)}</TableCell>
                <TableCell></TableCell>
              </TableRow>
            ))}
          </TableBody>
          <TableFooter>
            <TableRow>
              <TablePagination
                  rowsPerPageOptions={[50, 100, 150]}
                  onChangePage={function(event, number) { return }}
                  page={0}
                  count={0}
                  labelDisplayedRows={({ from, to, count }) => ""}
                  rowsPerPage={this.props.pageSize}
                  onChangeRowsPerPage={(event) => this.props.setPageSize(event)}
                  ActionsComponent={(props) => TaskTablePaginationActions(this.props)}
              />
            </TableRow>
          </TableFooter>
        </Table>
      </Paper>
    );
  }
}

TaskTableRaw.propTypes = {
  classes: PropTypes.object.isRequired,
  tasks: PropTypes.array.isRequired,
  nextPageToken: PropTypes.string.isRequired,
  prevPageToken: PropTypes.array.isRequired,
  setPageSize: PropTypes.func.isRequired,
  prevPage: PropTypes.func.isRequired,
  nextPage: PropTypes.func.isRequired,
};

function TaskTablePaginationActions(props) {
  //console.log("TaskTablePaginationActions props:", props)

  function handleBackButtonClick(event) {
    props.prevPage();
  }

  function handleNextButtonClick(event) {
    props.nextPage();
  }

  return (
    <div style={{flexShrink: 0}}>
      <IconButton 
       onClick={handleBackButtonClick} 
       disabled={props.prevPageToken.length === 0}
       aria-label="Previous Page"
      >
        <KeyboardArrowLeft />
      </IconButton>
      <IconButton
        onClick={handleNextButtonClick}
        disabled={props.nextPageToken === ""}
        aria-label="Next Page"
      >
        <KeyboardArrowRight />
      </IconButton>
    </div>
  );
}

TaskTablePaginationActions.propTypes = {
  nextPageToken: PropTypes.string.isRequired,
  prevPageToken: PropTypes.array.isRequired,
  prevPage: PropTypes.func.isRequired,
  nextPage: PropTypes.func.isRequired,
};

const TaskTable = withStyles(styles)(TaskTableRaw);
const NodeTable = withStyles(styles)(NodeTableRaw);
export { TaskTable, NodeTable };
