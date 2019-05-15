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

function TablePaginationActions(props) {
  const { count, page, rowsPerPage, onChangePage } = props;

  function handleBackButtonClick(event) {
    onChangePage(event, page - 1);
  }

  function handleNextButtonClick(event) {
    onChangePage(event, page + 1);
  }

  return (
    <div style={{flexShrink: 0}}>
      <IconButton onClick={handleBackButtonClick} disabled={page === 0} aria-label="Previous Page">
        <KeyboardArrowLeft />
      </IconButton>
      <IconButton
        onClick={handleNextButtonClick}
        disabled={page >= Math.ceil(count / rowsPerPage) - 1}
        aria-label="Next Page"
      >
        <KeyboardArrowRight />
      </IconButton>
    </div>
  );
}

TablePaginationActions.propTypes = {
  count: PropTypes.number.isRequired,
  onChangePage: PropTypes.func.isRequired,
  page: PropTypes.number.isRequired,
  rowsPerPage: PropTypes.number.isRequired,
};

let id = 0;
function createData(name, calories, fat, carbs, protein) {
  id += 1;
  return { id, name, calories, fat, carbs, protein };
}

const data = [
  createData('Frozen yoghurt', 159, 6.0, 24, 4.0),
  createData('Ice cream sandwich', 237, 9.0, 37, 4.3),
  createData('Eclair', 262, 16.0, 24, 6.0),
  createData('Cupcake', 305, 3.7, 67, 4.3),
  createData('Gingerbread', 356, 16.0, 49, 3.9),
  createData('Gingerbread', 356, 16.0, 49, 3.9),
  createData('Gingerbread', 356, 16.0, 49, 3.9),
  createData('Gingerbread', 356, 16.0, 49, 3.9),
];

function SimpleTableRaw(props) {
  const { classes } = props;

  const [page, setPage] = React.useState(0);
  const [rowsPerPage, setRowsPerPage] = React.useState(5);
  const emptyRows = rowsPerPage - Math.min(rowsPerPage, data.length - page * rowsPerPage);

  function handleChangePage(event, newPage) {
    setPage(newPage);
  }

  function handleChangeRowsPerPage(event) {
    setRowsPerPage(event.target.value);
  }

  return (
    <Paper className={classes.root}>
      <Table className={classes.table}>
        <TableHead>
          <TableRow>
            <TableCell>Dessert (100g serving)</TableCell>
            <TableCell align="right">Calories</TableCell>
            <TableCell align="right">Fat (g)</TableCell>
            <TableCell align="right">Carbs (g)</TableCell>
            <TableCell align="right">Protein (g)</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {data.slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage).map(n => (
            <TableRow key={n.id}>
              <TableCell component="th" scope="row">
                {n.name}
              </TableCell>
              <TableCell align="right">{n.calories}</TableCell>
              <TableCell align="right">{n.fat}</TableCell>
              <TableCell align="right">{n.carbs}</TableCell>
              <TableCell align="right">{n.protein}</TableCell>
            </TableRow>
          ))}
          {emptyRows > 0 && (
            <TableRow style={{ height: 48 * emptyRows }}>
              <TableCell colSpan={6} />
            </TableRow>
          )}
        </TableBody>
        <TableFooter>
          <TableRow>
            <TablePagination
                rowsPerPageOptions={[5, 10, 25]}
                count={data.length}
                rowsPerPage={rowsPerPage}
                page={page}
                onChangePage={handleChangePage}
                onChangeRowsPerPage={handleChangeRowsPerPage}
                ActionsComponent={TablePaginationActions}
            />
          </TableRow>
        </TableFooter>
      </Table>
    </Paper>
  );
}

SimpleTableRaw.propTypes = {
  classes: PropTypes.object.isRequired,
};

class NodeTableRaw extends React.Component {  
  render() {
    const { classes } = this.props;
    console.log("NodeTable props:", this.props)
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
                <TableCell><a href={"/v1/nodes/" + n.id}>{ n.hostname || n.id}</a></TableCell>
                <TableCell>{n.state}</TableCell>
                <TableCell>{n.task_ids.lengthe}</TableCell>
                <TableCell></TableCell>
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
      if (task.logs[0].start_time) {
        var now = new Date();
        if (this.isDone(task)) {
          if (task.logs[0].end_time) {
            now = Date.parse(task.logs[0].end_time);
          } else {
            return "--";
          }
        }
        var started = Date.parse(task.logs[0].start_time);
        var elapsed = now - started;
        return this.formatElapsedTime(elapsed);
      }
    }
    return "--";
  }

  creationTime(task) {
    if (task.creation_time) {
      var created = new Date(task.creation_time);
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
    const { classes } = this.props;
    const tasks = [
      {id: "bij587vpbjg2fb6m0beg",
       state: "CANCELED",
       name: "sh -c 'echo 1 && sleep 240'",
       executors:[{image: "alpine",
                   command: ["sh","-c","echo 1 && sleep 240"]}],
       logs: [{logs: [{start_time: "2019-04-04T11:59:43.977367-07:00",
                       stdout: "1\n"}],
               metadata:{hostname: "BICB230"},
               start_time: "2019-04-04T11:59:43.854152-07:00"}],
       creation_time: "2019-04-04T11:59:43.793262-07:00"},
    ]

    console.log("TaskTable props:", this.props)

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
            {tasks.map(t => (
              <TableRow key={t.id}>
                <TableCell><a href={"/v1/tasks/" + t.id}>{t.id}</a></TableCell>
                <TableCell>{t.state}</TableCell>
                <TableCell>{t.name}</TableCell>
                <TableCell>{this.creationTime(t)}</TableCell>
                <TableCell>{this.elapsedTime(t)}</TableCell>
                <TableCell></TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Paper>
    );
  }
}

TaskTableRaw.propTypes = {
  classes: PropTypes.object.isRequired,
  tasks: PropTypes.array.isRequired,
  prevPage: PropTypes.func.isRequired,
  nextPage: PropTypes.func.isRequired,
};

const TaskTable = withStyles(styles)(TaskTableRaw);
const NodeTable = withStyles(styles)(NodeTableRaw);
const SimpleTable = withStyles(styles)(SimpleTableRaw);
export { TaskTable, NodeTable, SimpleTable };
