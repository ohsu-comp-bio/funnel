import React from 'react';
import PropTypes from 'prop-types';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableRow from '@material-ui/core/TableRow';


class SystemInfo extends React.Component {

  renderStateCounts() {
    const cellstyle = {
      fontSize: '10pt'
    }
    if (this.props.info.taskStateCounts === undefined) {
      return (
        <TableRow key="doc">
          <TableCell style={cellstyle}><b>Task State Counts</b></TableCell>
          <TableCell style={cellstyle}>Not Available</TableCell>
        </TableRow>
      )
    } else {
      return (
        <TableRow key="doc">
          <TableCell style={cellstyle}><b>Task State Counts</b></TableCell>
          {Object.keys(this.props.info.taskStateCounts).map(k => (
              <TableRow key={k}>
              <TableCell style={cellstyle}>{k}</TableCell>
              <TableCell style={cellstyle}>{this.props.info.taskStateCounts[k]}</TableCell>
              </TableRow>
          ))}
        </TableRow>
      )
    }
  }

  render() {
    const cellstyle = {
      fontSize: '10pt'
    }
    return (
        <Table>
          <TableBody>
              <TableRow key="name">
                <TableCell style={cellstyle}><b>Name</b></TableCell>
                <TableCell style={cellstyle}>{this.props.info.name}</TableCell>
              </TableRow>
              <TableRow key="doc">
                <TableCell style={cellstyle}><b>Doc</b></TableCell>
                <TableCell style={cellstyle}><pre>{this.props.info.doc}</pre></TableCell>
              </TableRow>
              {this.renderStateCounts()}
          </TableBody>
        </Table>
    );
  }
}

SystemInfo.propTypes = {
  info: PropTypes.object.isRequired,
};

export { SystemInfo };
