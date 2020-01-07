import React from 'react';
import PropTypes from 'prop-types';
import ListItem from '@material-ui/core/ListItem';
import ListItemText from '@material-ui/core/ListItemText';
import ListSubheader from '@material-ui/core/ListSubheader';
import Input from '@material-ui/core/Input';
import InputLabel from '@material-ui/core/InputLabel';
import MenuItem from '@material-ui/core/MenuItem';
import FormGroup from '@material-ui/core/FormGroup';
import Select from '@material-ui/core/Select';
import TextField from '@material-ui/core/TextField';
import Icon from '@material-ui/core/Icon';
import IconButton from '@material-ui/core/IconButton';
import DeleteIcon from '@material-ui/icons/Delete';
import { Link } from "react-router-dom";

function ListItemLink(props) {
  return <ListItem button component={Link} {...props} />;
}

const NavListItems = (
  <div>
    <ListItemLink to="/tasks">
      <ListItemText primary="Tasks" />
    </ListItemLink>
    <ListItemLink to="/nodes">
      <ListItemText primary="Nodes" />
    </ListItemLink>
    <ListItemLink to="/service-info">
      <ListItemText primary="Service Info" />
    </ListItemLink>
  </div>
);

function TaskFilters({show, stateFilter, tagsFilter, setStateFilter, setTagsFilter}) {

  const updateTagFilter = (index, newTagValue) => {
    let tags =[...tagsFilter];
    tags[index] = newTagValue;
    setTagsFilter(tags);
  };

  const addTagFilter = () => {
    const tags = tagsFilter.concat({key: "", value: ""});
    setTagsFilter(tags);
  };

  const removeTagFilter = (index) => {
    const tags = tagsFilter.filter((item, j) => index !== j);
    setTagsFilter(tags);
  };

  if (show === false) {
    return (<div />);
  };
  return (
    <div>
      <ListSubheader>Task Filters</ListSubheader>
      <ListItem style={{paddingBottom:"0"}}>
        <InputLabel shrink htmlFor="state-placeholder">
          State
        </InputLabel>
      </ListItem>
      <ListItem style={{paddingTop:"5px"}}>
        <Select
          value={stateFilter}
          onChange={(event) => setStateFilter(event.target.value)}
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
      </ListItem>
      <ListItem style={{paddingBottom:"0"}}>
        <InputLabel shrink htmlFor="state-placeholder">
          Tags
        </InputLabel>
        <IconButton onClick={addTagFilter}>
          <Icon >add_circle</Icon>
        </IconButton>
      </ListItem>
      {tagsFilter.map((tag, index) => (
        <ListItem key={index} style={{paddingTop:"5px", paddingRight:"0px"}}>
          <TagFilter index={index} tag={tag} updateTagFilter={updateTagFilter} />
          <IconButton onClick={() => removeTagFilter(index)}>
            <DeleteIcon />
          </IconButton>
        </ListItem>
      ))}
    </div>
  );
};

TaskFilters.propTypes = {
  show: PropTypes.bool.isRequired,
  stateFilter: PropTypes.string.isRequired,
  tagsFilter: PropTypes.array.isRequired,
  setStateFilter: PropTypes.func.isRequired,
  setTagsFilter: PropTypes.func.isRequired,
};

const useDebounce = (value, delay) => {
  const [debouncedValue, setDebouncedValue] = React.useState(value);

  React.useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(value);
    }, delay);
    
    return () => {
      clearTimeout(handler);
    };
  }, [value, delay]);
  
  return debouncedValue;
};

function TagFilter({tag, index, updateTagFilter}) {
  const [key, setKey] = React.useState(tag.key);
  const [value, setValue] = React.useState(tag.value);
  const debouncedKey = useDebounce(key, 500);
  const debouncedValue = useDebounce(value, 500);

  React.useEffect(() => {
    updateTagFilter(index, {"key": debouncedKey, "value": debouncedValue});
  }, [debouncedKey, debouncedValue, index]);

  return(
    <div style={{borderColor: "#eeeeee", borderWidth: "1px", borderStyle:"solid"}}>
      <form autoComplete="off">
        <FormGroup style={{flexDirection:"row", padding:"10px"}}>
          <TextField 
            id="key"
            label="Key"
            placeholder=""
            value={key}
            margin="normal"
            style={{marginTop:"0px"}}
            onChange={(event) => setKey(event.target.value)}
          />
          <TextField 
            id="value"
            label="Value"
            placeholder=""
            value={value}
            margin="normal"
            style={{marginTop:"0px"}}
            onChange={(event) => setValue(event.target.value)}
          />
        </FormGroup>
      </form>
    </div>
  );
};

TagFilter.propTypes = {
  tag: PropTypes.object.isRequired,
  index: PropTypes.number.isRequired,
  updateTagFilter: PropTypes.func.isRequired,
};

export {NavListItems, TaskFilters};
