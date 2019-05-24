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
import debounce from 'lodash/debounce'
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

class TaskFilters extends React.Component {
 
  render() {
    //console.log("TaskFilters props:", this.props)
    if (this.props.show === false) {
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
            value={this.props.stateFilter}
            onChange={(event) => this.props.updateState(event)}
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
          <IconButton onClick={this.props.addTag}>
            <Icon >add_circle</Icon>
          </IconButton>
        </ListItem>
        {this.props.tagsFilter.map((tag, index) => (
          <ListItem key={index} style={{paddingTop:"5px", paddingRight:"0px"}}>
            <TagFilter index={index} tag={tag} updateTags={this.props.updateTags} />
            <IconButton onClick={() => this.props.removeTag(index)}>
              <DeleteIcon />
            </IconButton>
          </ListItem>
        ))}
      </div>
    );
  };
};

TaskFilters.propTypes = {
  show: PropTypes.bool.isRequired,
  stateFilter: PropTypes.string.isRequired,
  tagsFilter: PropTypes.array.isRequired,
  updateState: PropTypes.func.isRequired,
  updateTags: PropTypes.func.isRequired,
  addTag: PropTypes.func.isRequired,
};

class TagFilter extends React.Component {
  constructor(props) {
    super(props)
    this.state = {
      key: this.props.tag.key, 
      value: this.props.tag.value,
    };
    this.updateTags = debounce(this.props.updateTags, 500)
  };

  update = (event) => {
    this.setState({[event.target.id]: event.target.value});
    this.updateTags(this.props.index, event.target.id, event.target.value);
  };

  render() {
    //console.log("TagFilter props:", this.props)
    //console.log("TagFilter state:", this.state)
    return(
      <div style={{borderColor: "#eeeeee", borderWidth: "1px", borderStyle:"solid"}}>
        <form autoComplete="off">
          <FormGroup style={{flexDirection:"row", padding:"10px"}}>
            <TextField 
             id="key"
             label="Key"
             placeholder=""
             value={this.state.key}
             margin="normal"
             style={{marginTop:"0px"}}
             onChange={(event) => this.update(event)}
            />
            <TextField 
             id="value"
             label="Value"
             placeholder=""
             value={this.state.value}
             margin="normal"
             style={{marginTop:"0px"}}
             onChange={(event) => this.update(event)}
            />
          </FormGroup>
        </form>
      </div>
    );
  };
}

TagFilter.propTypes = {
  tag: PropTypes.object.isRequired,
  index: PropTypes.number.isRequired,
  updateTags: PropTypes.func.isRequired,
};

export {NavListItems, TaskFilters};
