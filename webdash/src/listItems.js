import React from 'react';
import ListItem from '@material-ui/core/ListItem';
import ListItemText from '@material-ui/core/ListItemText';
import ListSubheader from '@material-ui/core/ListSubheader';
import Input from '@material-ui/core/Input';
import InputLabel from '@material-ui/core/InputLabel';
import MenuItem from '@material-ui/core/MenuItem';
import FormControl from '@material-ui/core/FormControl';
import Select from '@material-ui/core/Select';

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

export const filterListItems = (
  <div>
    <ListSubheader>Filters</ListSubheader>
    <ListItem>
    <form autoComplete="off">
      <FormControl>
        <InputLabel shrink htmlFor="state-placeholder">
          State
        </InputLabel>
        <Select
          value={""}
          //onChange={}
          input={<Input name="state" id="state-placeholder" />}
          displayEmpty
          autoWidth
          name="state"
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
  </div>
);