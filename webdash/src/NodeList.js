import React from 'react';
import PropTypes from 'prop-types';
import { withStyles } from '@material-ui/core/styles';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';
import Paper from '@material-ui/core/Paper';

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
              <TableCell>Hostname</TableCell>
              <TableCell>State</TableCell>
              <TableCell>Tasks</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {this.props.nodes.map(n => (
              <TableRow hover key={n.id}>
                <TableCell><a href={"/nodes/" + n.id}>{n.id}</a></TableCell>
                <TableCell>{ n.hostname }</TableCell>
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

const NodeTable = withStyles(styles)(NodeTableRaw);
export { NodeTable };
