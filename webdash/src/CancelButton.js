import React from 'react';
import Button from '@material-ui/core/Button';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogContentText from '@material-ui/core/DialogContentText';
import DialogTitle from '@material-ui/core/DialogTitle';

export default function CancelButton(task) {
  const [open, setOpen] = React.useState(false);

  const cancelTask = () => {
    var url = new URL("/v1/tasks/" + task.id + ":cancel", window.location.origin);
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

  const handleClickOpen = () => {
    setOpen(true);
  };

  const handleNo = () => {
    setOpen(false);
  };

  const handleYes = () => {
    cancelTask();
    setOpen(false);
  };

  return (
    <div>
      <Button variant="outlined" onClick={handleClickOpen}>
        Cancel
      </Button>
      <Dialog
        open={open}
        onClose={null}
        aria-labelledby="alert-dialog-title"
        aria-describedby="alert-dialog-description"
      >
        <DialogTitle id="alert-dialog-title">Cancel task: {task.id}?</DialogTitle>
        <DialogContent>
          <DialogContentText id="alert-dialog-description">
            This action can not be undone.
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleNo} color="primary">
            No
          </Button>
          <Button onClick={handleYes} color="primary" autoFocus>
            Yes
          </Button>
        </DialogActions>
      </Dialog>
    </div>
  );
};

