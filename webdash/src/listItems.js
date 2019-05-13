import React from 'react';
import ListItem from '@material-ui/core/ListItem';
import ListItemText from '@material-ui/core/ListItemText';
import ListSubheader from '@material-ui/core/ListSubheader';
import Input from '@material-ui/core/Input';
import InputLabel from '@material-ui/core/InputLabel';
import MenuItem from '@material-ui/core/MenuItem';
import FormControl from '@material-ui/core/FormControl';
import Select from '@material-ui/core/Select';
import TextField from '@material-ui/core/TextField';

function ListItemLink(props) {
  return <ListItem button component="a" {...props} />;
}

export const mainListItems = (
  <div>
    <ListItemLink href="/tasks">
       <ListItemText primary="Tasks" />
    </ListItemLink>
    <ListItemLink href="/nodes">
      <ListItemText primary="Nodes" />
    </ListItemLink>
    <ListItemLink href="/service-info">
       <ListItemText primary="Service Info" />
    </ListItemLink>
  </div>
);

class FilterList extends React.Component {

  render() {
    console.log("FilterList props:", this.props)
    if (typeof this.props.show === 'boolean' && this.props.show === true) {
      return (
      <div>
        <ListSubheader>Filters</ListSubheader>
        <ListItem>
        <form autoComplete="off">
          <FormControl>
            <InputLabel shrink htmlFor="state-placeholder">
              State
            </InputLabel>
            <Select
              value={this.props.stateFilter}
              onChange={(event) => this.props.updateFn(event)}
              input={<Input name="stateFilter" id="state-placeholder" />}
              displayEmpty
              autoWidth
              name="stateFilter"
            >
              <MenuItem value="">
                <em>None</em>
              </MenuItem>
              <MenuItem value="QUEUED">Queued</MenuItem>
              <MenuItem value="INITIALIZING">Initializing</MenuItem>
              <MenuItem value="RUNNING">Running</MenuItem>
              <MenuItem value="COMPLETE">Complete</MenuItem>
              <MenuItem value="CANCELED">Canceled</MenuItem>
              <MenuItem value="EXECUTOR_ERROR">Executor Error</MenuItem>
              <MenuItem value="SYSTEM_ERROR">System error</MenuItem>
            </Select>
          </FormControl>
        </form>
        </ListItem>
        <ListItem>
        <form autoComplete="off">
          <FormControl>
            <InputLabel shrink htmlFor="state-placeholder">
              Tags
            </InputLabel>
            <br />
            <TextField
              id="tag-key"
              label="Key"
              placeholder=""
              fullWidth
              margin="normal"
              InputLabelProps={{
                shrink: true,
              }}
            />
            <TextField
              id="tag-value"
              label="Value"
              placeholder=""
              fullWidth
              margin="normal"
              InputLabelProps={{
                shrink: true,
              }}
            />
          </FormControl>
        </form>
        </ListItem>
      </div>
      )
    } else {
      return (<div />)
    }
  }
}

export {FilterList};
