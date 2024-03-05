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
        <TableRow className={classes.row} key="contactUrl">
          <TableCell className={classNames(classes.cell, classes.key)}><b>contactUrl</b></TableCell>
          <TableCell className={classNames(classes.cell, classes.value)}><pre>{info.contactUrl}</pre></TableCell>
        </TableRow>
        <TableRow className={classes.row} key="createdAt">
          <TableCell className={classNames(classes.cell, classes.key)}><b>createdAt</b></TableCell>
          <TableCell className={classNames(classes.cell, classes.value)}><pre>{info.createdAt}</pre></TableCell>
        </TableRow>
        <TableRow className={classes.row} key="description">
          <TableCell className={classNames(classes.cell, classes.key)}><b>description</b></TableCell>
          <TableCell className={classNames(classes.cell, classes.value)}><pre>{info.description}</pre></TableCell>
        </TableRow>
        <TableRow className={classes.row} key="documentationUrl">
          <TableCell className={classNames(classes.cell, classes.key)}><b>documentationUrl</b></TableCell>
          <TableCell className={classNames(classes.cell, classes.value)}><pre>{info.documentationUrl}</pre></TableCell>
        </TableRow>
        <TableRow className={classes.row} key="id">
          <TableCell className={classNames(classes.cell, classes.key)}><b>id</b></TableCell>
          <TableCell className={classNames(classes.cell, classes.value)}><pre>{info.id}</pre></TableCell>
        </TableRow>
        <TableRow className={classes.row} key="organization">
          <TableCell className={classNames(classes.cell, classes.key)}><b>organization</b></TableCell>
          <TableBody>
            <TableRow className={classes.row}>
              <TableCell className={classNames(classes.cell, classes.value)}><pre>{info.organization?.name}</pre></TableCell>
            </TableRow>
            <TableRow>
              <TableCell className={classNames(classes.cell, classes.value)}><pre>{info.organization?.url}</pre></TableCell>
            </TableRow>
          </TableBody>
        </TableRow>
        <TableRow className={classes.row} key="storage">
          <TableCell className={classNames(classes.cell, classes.key)}><b>storage</b></TableCell>
          <TableCell className={classNames(classes.cell, classes.value)}><pre>{info.storage}</pre></TableCell>
        </TableRow>
        <TableRow className={classes.row} key="tesResources_backend_parameter">
          <TableCell className={classNames(classes.cell, classes.key)}><b>tesResources_backend_parameter</b></TableCell>
          <TableCell className={classNames(classes.cell, classes.value)}><pre>{info.tesResources_backend_parameter}</pre></TableCell>
        </TableRow>
        <TableRow className={classes.row} key="type">
          <TableCell className={classNames(classes.cell, classes.key)}><b>type</b></TableCell>
          <TableBody>
            <TableRow className={classes.row}>
              <TableCell className={classNames(classes.cell, classes.value)}><pre>{info.type?.artifact}</pre></TableCell>
            </TableRow>
            <TableRow>
              <TableCell className={classNames(classes.cell, classes.value)}><pre>{info.type?.group}</pre></TableCell>
            </TableRow>
            <TableRow>
              <TableCell className={classNames(classes.cell, classes.value)}><pre>{info.type?.version}</pre></TableCell>
            </TableRow>
          </TableBody>
        </TableRow>
        <TableRow className={classes.row} key="updatedAt">
          <TableCell className={classNames(classes.cell, classes.key)}><b>updatedAt</b></TableCell>
          <TableCell className={classNames(classes.cell, classes.value)}><pre>{info.updatedAt}</pre></TableCell>
        </TableRow>
        <TableRow className={classes.row} key="version">
          <TableCell className={classNames(classes.cell, classes.key)}><b>version</b></TableCell>
          <TableCell className={classNames(classes.cell, classes.value)}><pre>{info.version}</pre></TableCell>
        </TableRow>
      </TableBody>
    </Table>
  );
};

SystemInfoRaw.propTypes = {
  info: PropTypes.object.isRequired,
};

const SystemInfo = withStyles(styles)(SystemInfoRaw);
export { SystemInfo };
