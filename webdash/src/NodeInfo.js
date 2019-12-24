import React from 'react';
import PropTypes from 'prop-types';
import { withStyles } from '@material-ui/core/styles';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableRow from '@material-ui/core/TableRow';
import classNames from 'classnames';
import { Link } from "react-router-dom";

import { formatTimestamp } from './utils';
import { styles } from './TaskInfo';


function NodeInfoRaw(props) {
  const { classes, node } = props;

  const renderRow = (key, val, formatFunc) => {
    if ( val ) {
      if (typeof formatFunc === "function") {
        val = formatFunc(val);
      } else if (formatFunc) {
        console.log("renderRow: formatFunc was not a function:", typeof(formatFunc));
      }
    }
    if ( key && val ) {
      return (
        <TableRow key={key} className={classes.row}>
          <TableCell className={classNames(classes.cell, classes.key)}><b>{key}</b></TableCell>
          <TableCell className={classNames(classes.cell, classes.value)}>{val}</TableCell>
        </TableRow>
      );
    }
    return null;
  };

  const renderTasks = (taskList) => {
    if (!taskList) {
      return null;
    }
    return (
      renderRow('Tasks', taskList.map(tid => ( 
          <div><Link to={"/tasks/" + tid}>{tid}</Link><br/></div>
      )))
    );
  };

  const resourceString = (resources) => {
    if ( resources ) {
      const r = resources;
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
      return s;
    }
    return null;
  };

  const renderNode = (node) => {
    if (!node) {
      return null;
    }
    return (
      <div>
        <Table className={classes.table}>
          <TableBody>
            {renderRow('ID', node.id)}
            {renderRow('Hostname', node.hostname)}
            {renderRow('State', node.state)}
            {renderRow('Resources', node.resources, resourceString)}
            {renderRow('Available', node.available, resourceString)}
            {renderRow('Last Ping', node.lastPing, formatTimestamp)}
            {renderRow('Version', node.version)}
            {renderTasks(node.taskIds)}
          </TableBody>
        </Table>
      </div>
    );
  };

  return (
    <div>
      {renderNode(node)}
    </div>
  );
}

NodeInfoRaw.propTypes = {
  node: PropTypes.object.isRequired,
};

const NodeInfo = withStyles(styles)(NodeInfoRaw);
export { NodeInfo };
