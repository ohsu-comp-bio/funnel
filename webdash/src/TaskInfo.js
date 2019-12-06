import React from 'react';
import PropTypes from 'prop-types';
import { withStyles } from '@material-ui/core/styles';
import Divider from '@material-ui/core/Divider';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableRow from '@material-ui/core/TableRow';
import Typography from '@material-ui/core/Typography';
import classNames from 'classnames';

import { formatDate, elapsedTime } from './utils';

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

class TaskInfoRaw extends React.Component {

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

  renderResources(resources) {
    const { classes } = this.props;
    if ( resources ) {
      const r = resources
      var s = r.cpuCores + " CPU cores";
      if (r.ramGb) {
        s += ", " + r.ramGb + " GB RAM";
      }
      if (r.diskGb) {
        s += ", " + r.diskGb + " GB disk space";
      }
      if (r.preemptible) {
        s += ", preemptible";
      }
      return (
        <TableRow key='Resources' className={classes.row}>
          <TableCell className={classNames(classes.cell, classes.key)}><b>Resources</b></TableCell>
          <TableCell className={classNames(classes.cell, classes.value)}>{s}</TableCell>
        </TableRow>
      )
    }
    return
  }

  renderTitle(title) {
    if (title) {
      return (
        <div>
          <Typography variant="h6">{title}</Typography>
          <Divider />
        </div>
      )
    }
    return
  }

  renderKV(data, title, defaultPadding='40px 0px 0px 0px') {
    const { classes } = this.props;
    if ( data ) {
      return (
        <div style={{padding: defaultPadding}}>
          {this.renderTitle(title)}
          <Table className={classes.table}>
            <TableBody>
              {Object.keys(data).map(k => (
               <TableRow key={k} className={classes.row}>
                 <TableCell className={classNames(classes.cell, classes.key)}>
                   <b>{k}</b>
                 </TableCell>
                 <TableCell className={classNames(classes.cell, classes.value)}>
                   {data[k]}
                 </TableCell>
               </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )
    }
    return
  }

  // should we truncate content???
  renderFileArrays(files, title) {
    if ( files && Array.isArray(files)) {
      return (
        <div style={{padding: '40px 0px 0px 0px'}}>
          {this.renderTitle(title)}
          {files.map(item => (
            this.renderKV(item, null, '20px 0px 0px 0px')
          ))}
        </div>
      )
    }
    return
  }

  renderMetadata(logs) {
    if ( logs && logs.length ) {
      return this.renderKV(logs[0].metadata, "Metadata")
    }
    return
  }

  renderOutputFileLog(logs) {
    if ( logs && logs.length ) {
      return this.renderFileArrays(logs[0].outputs, "Output File Log")
    }
    return
  }

  // should we truncate stdout / stderr ???
  renderExecutors(task) {
    const { classes } = this.props;
    if ( task.executors ) {
      var executors = task.executors
      var logs = [{}]
      if ( task.logs && task.logs && task.logs[0].logs ) {
        logs = task.logs[0].logs
      }
      const cmdString = function(cmd) {
       return cmd.join(" ") 
      }
      const preFormat = function(s) {
        return <pre>{s}</pre>
      }
      return (
        <div style={{padding: '40px 0px 0px 0px'}}>
          {this.renderTitle('Executors')}
          {executors.map((exec, index) => (
            <Table className={classes.table}>
              <TableBody>
                <TableRow key='Cmd' className={classes.row}>
                  <TableCell className={classNames(classes.cell, classes.key)}><b>Command</b></TableCell>
                  <TableCell className={classNames(classes.cell, classes.value)}>{cmdString(exec.command)}</TableCell>
                </TableRow>
                <TableRow key='Image' className={classes.row}>
                  <TableCell className={classNames(classes.cell, classes.key)}><b>Image</b></TableCell>
                  <TableCell className={classNames(classes.cell, classes.value)}>{exec.image}</TableCell>
                </TableRow>
                {this.renderRow('Workdir', exec.workdir)}
                {this.renderRow('Env', this.renderKV(exec.env, null, '0px 0px 0px 0px'))}
                {this.renderRow('StartTime', logs[index].startTime, formatDate)}
                {this.renderRow('EndTime', logs[index].endTime, formatDate)}
                {this.renderRow('Exit Code', logs[index].exitCode)}
                {this.renderRow('Stdout', logs[index].stdout, preFormat)}
                {this.renderRow('Stderr', logs[index].stderr, preFormat)}
              </TableBody>
            </Table>
          ))}
        </div>
      )
    }
    return
  }

  renderSysLogs(logs) {
    if ( logs && logs.length && logs[0].systemLogs && Array.isArray(logs[0].systemLogs)) {
      const syslogs = logs[0].systemLogs
      var entryList = syslogs.map(item => {
        var parts = item.split("' ").map(p => p.replace(/'/g, "").split("="))
        return parts
      })
      var entries = []
      for (var i in entryList) {        
        var entry = {}
        for (var j in entryList[i]) {
          if (entryList[i][j][1] !== "") {
            entry[entryList[i][j][0]] = entryList[i][j][1]
          }
        }
        entries.push(entry)
      }
      return (
        <div style={{padding: '40px 0px 0px 0px'}}>
          {this.renderTitle('System Logs')}
          {entries.map(e => this.renderKV(e, null, '20px 0px 0px 0px'))}
        </div>
      )
    }
    return
  }

  renderTask(task) {
    const { classes } = this.props;
    if (!task) {
      return
    }
    return(
      <div>
        <Table className={classes.table}>
          <TableBody>
            {this.renderRow('Name', task.name)}
            {this.renderRow('ID', task.id)}
            {this.renderRow('State', task.state)}
            {this.renderRow('Description', task.description)}
            {this.renderResources(task.resources)}
            {this.renderRow('Creation Time', task.creationTime, formatDate)}
            {this.renderRow('Elapsed Time', task, elapsedTime)}
          </TableBody>
        </Table>

        {/* Tags */}
        {this.renderKV(task.tags, 'Tags')}

        {/* Metadata */}
        {this.renderMetadata(task.logs, 'Metadata')}

        {/* Inputs */}
        {this.renderFileArrays(task.inputs, 'Inputs')}

        {/* Outputs */}
        {this.renderFileArrays(task.outputs, 'Outputs')}

        {/* Executors */}
        {this.renderExecutors(task)}

        {/* Output File Logs */}
        {this.renderOutputFileLog(task.logs, 'Output File Log')}

        {/* System Logs */}
        {this.renderSysLogs(task.logs, 'System Logs')}

      </div>
    )
  }

  render() {
    const task = this.props.task
    return (
      <div>
        {this.renderTask(task)}
      </div>
    )
  }
}

TaskInfoRaw.propTypes = {
  task: PropTypes.object.isRequired,
};

const TaskInfo = withStyles(styles)(TaskInfoRaw);
export { TaskInfo };
