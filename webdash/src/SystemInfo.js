import React from 'react';
import PropTypes from 'prop-types';
import { withStyles } from '@material-ui/core/styles';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableRow from '@material-ui/core/TableRow';
import classNames from 'classnames';

import { styles } from './TaskInfo';

function SystemInfoRaw(props) {
  const { classes, info } = props;

  const renderStateCounts = () => {
    if (info.taskStateCounts === undefined) {
      return (
        <TableRow className={classes.row} key="counts">
          <TableCell className={classNames(classes.cell, classes.key)}><b>Task State Counts</b></TableCell>
          <TableCell className={classNames(classes.cell, classes.value)}>Not Available</TableCell>
        </TableRow>
      );
    } else {
      return (
        <TableRow className={classes.row} key="counts">
          <TableCell className={classNames(classes.cell, classes.key)}><b>Task State Counts</b></TableCell>
          <TableCell className={classNames(classes.cell, classes.value)}>
            <Table>
              <TableBody>
                {Object.keys(info.taskStateCounts).map(k => (
                  <TableRow className={classes.row} key={k}>
                    <TableCell className={classNames(classes.cell, classes.key)}>{k}</TableCell>
                    <TableCell className={classNames(classes.cell, classes.value)}>{info.taskStateCounts[k]}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableCell>
        </TableRow>
      );
    }
  };

  return (
    <Table className={classes.table}>
      <TableBody>
        <TableRow className={classes.row} key="name">
          <TableCell className={classNames(classes.cell, classes.key)}><b>Name</b></TableCell>
          <TableCell className={classNames(classes.cell, classes.value)}>{info.name}</TableCell>
        </TableRow>
        <TableRow className={classes.row} key="doc">
          <TableCell className={classNames(classes.cell, classes.key)}><b>Doc</b></TableCell>
          <TableCell className={classNames(classes.cell, classes.value)}><pre>{info.doc}</pre></TableCell>
        </TableRow>
        {renderStateCounts()}
      </TableBody>
    </Table>
  );
};

SystemInfoRaw.propTypes = {
  info: PropTypes.object.isRequired,
};

const SystemInfo = withStyles(styles)(SystemInfoRaw);
export { SystemInfo };
