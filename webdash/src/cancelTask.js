import React from 'react';
import Button from '@material-ui/core/Button';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogContentText from '@material-ui/core/DialogContentText';
import DialogTitle from '@material-ui/core/DialogTitle';

export default class CancelButton extends React.Component {
  constructor(props) {
    super(props)
    this.state = {
      open: false
    }
  }

  cancelTask(id) {
    var url = new URL("/v1/tasks/" + id + ":cancel", window.location.origin);
    //console.log("cancelTask url:", url);
    fetch(url.toString(), {
      method: 'POST',
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json',
      },
    })
    .then(response => response.json())
    .then(
      (result) => {
        return result;
      },
      (error) => {
        console.log("cancelTask", url.toString(), "error:", error);
        return undefined;
      },
    );
  };

  handleClickOpen = () => {
    this.setState({open: true});
  };

  handleNo = () => {
    console.log("ABORT CANCEL")
    this.setState({open: false});
  };

  handleYes = () => {
    console.log("CANCEL")
    this.cancelTask(this.props.task.id);
    this.setState({open: false});
  };

  render() {
    return (
      <div>
        <Button variant="outlined" onClick={this.handleClickOpen}>
          Cancel
        </Button>
        <Dialog
          open={this.state.open}
          onClose={this.handleClose}
          aria-labelledby="alert-dialog-title"
          aria-describedby="alert-dialog-description"
        >
          <DialogTitle id="alert-dialog-title">Cancel task: {this.props.task.id}?</DialogTitle>
          <DialogContent>
            <DialogContentText id="alert-dialog-description">
              This action can not be undone.
            </DialogContentText>
          </DialogContent>
          <DialogActions>
            <Button onClick={this.handleNo} color="primary">
              No
            </Button>
            <Button onClick={this.handleYes} color="primary" autoFocus>
              Yes
            </Button>
          </DialogActions>
        </Dialog>
      </div>
    );
  }
}
