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

import { formatDate, elapsedTime, renderCancelButton } from './utils';

const styles = {
  root: {
    width: '100%',
    overflowX: 'auto',
  },
  table: {
    minWidth: 700,
  },
};

class TaskTableRaw extends React.Component {

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
              <TableRow hover key={t.id}>
                <TableCell><a href={"/tasks/" + t.id}>{t.id}</a></TableCell>
                <TableCell>{t.state}</TableCell>
                <TableCell>{t.name}</TableCell>
                <TableCell>{formatDate(t.creationTime)}</TableCell>
                <TableCell>{elapsedTime(t)}</TableCell>
                <TableCell>{renderCancelButton(t)}</TableCell>
              </TableRow>
            ))}
          </TableBody>
          <TableFooter>
            <TableRow>
              <TablePagination
                  rowsPerPageOptions={[25, 50, 100, 250, 500]}
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
export { TaskTable };
