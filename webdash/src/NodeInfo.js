import React from 'react';
import PropTypes from 'prop-types';
import { withStyles } from '@material-ui/core/styles';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableRow from '@material-ui/core/TableRow';
import classNames from 'classnames';

import { formatTimestamp } from './utils';

const styles = {
  table: {
    width: '100%',
    overflowX: 'auto',
  },
  row: {
    height: '0',
  },
  cell: {
    fontSize: '10pt',
    borderBottomStyle: 'none',
    padding:'1px',
  },
  key: {
    width: '20%',
  },
  value: {
    width: '80%',
  },
};

class NodeInfoRaw extends React.Component {

  renderRow(key, val, formatFunc) {
    const { classes } = this.props;
    if ( val ) {
      if (typeof formatFunc === "function") {
        val = formatFunc(val)
      } else if (formatFunc) {
        console.log("renderRow: formatFunc was not a function:", typeof(formatFunc))
      }
    }
    if ( key && val ) {
      return (
        <TableRow key={key} className={classes.row}>
          <TableCell className={classNames(classes.cell, classes.key)}><b>{key}</b></TableCell>
          <TableCell className={classNames(classes.cell, classes.value)}>{val}</TableCell>
        </TableRow>
      )
    }
    return
  }

  renderTasks(taskList) {
    if (!taskList) {
      return
    }
    return (
      this.renderRow('Tasks', taskList.map(tid => ( 
          <div><a href={"/tasks/" + tid}>{tid}</a><br/></div>
      )))
    )
  }

  resourceString(resources) {
    if ( resources ) {
      const r = resources
      var s = r.cpus + " CPU cores";
      if (r.ramGb) {
        s += ", " + r.ramGb + " GB RAM";
      }
      if (r.diskGb) {
        s += ", " + r.diskGb + " GB disk space";
      }
      if (r.preemptible) {
        s += ", preemptible";
      }
      return s
    }
    return
  }

  renderNode(node) {
    const { classes } = this.props;
    if (!node) {
      return
    }
    return(
      <div>
        <Table className={classes.table}>
          <TableBody>
            {this.renderRow('ID', node.id)}
            {this.renderRow('Hostname', node.hostname)}
            {this.renderRow('State', node.state)}
            {this.renderRow('Resources', node.resources, this.resourceString)}
            {this.renderRow('Available', node.available, this.resourceString)}
            {this.renderRow('Last Ping', node.lastPing, formatTimestamp)}
            {this.renderRow('Version', node.version)}
            {this.renderTasks(node.taskIds)}
          </TableBody>
        </Table>
      </div>
    )
  }

  render() {
    const node = this.props.node
    return (
      <div>
        {this.renderNode(node)}
      </div>
    )
  }
}

NodeInfoRaw.propTypes = {
  node: PropTypes.object.isRequired,
};

const NodeInfo = withStyles(styles)(NodeInfoRaw);
export { NodeInfo };
