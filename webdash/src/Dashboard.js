import React from 'react';
import { Route, Redirect, Switch, withRouter } from "react-router-dom";
import PropTypes from 'prop-types';
import classNames from 'classnames';
import { withStyles } from '@material-ui/core/styles';
import CssBaseline from '@material-ui/core/CssBaseline';
import Drawer from '@material-ui/core/Drawer';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import List from '@material-ui/core/List';
import Typography from '@material-ui/core/Typography';
import Divider from '@material-ui/core/Divider';
import IconButton from '@material-ui/core/IconButton';
import MenuIcon from '@material-ui/icons/Menu';
import ChevronLeftIcon from '@material-ui/icons/ChevronLeft';

import { NavListItems, TaskFilters } from './DrawerFilters';
import { TaskList, Task, Node, NodeList, ServiceInfo, NoMatch } from './Pages';

const drawerWidth = 260;

const styles = theme => ({
  root: {
    display: 'flex',
  },
  toolbar: {
    paddingRight: 24, // keep right padding when drawer closed
  },
  toolbarIcon: {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'flex-end',
    padding: '0 8px',
    ...theme.mixins.toolbar,
  },
  appBar: {
    backgroundColor: "#000000",
    zIndex: theme.zIndex.drawer + 1,
    transition: theme.transitions.create(['width', 'margin'], {
      easing: theme.transitions.easing.sharp,
      duration: theme.transitions.duration.leavingScreen,
    }),
  },
  appBarShift: {
    marginLeft: drawerWidth,
    width: `calc(100% - ${drawerWidth}px)`,
    transition: theme.transitions.create(['width', 'margin'], {
      easing: theme.transitions.easing.sharp,
      duration: theme.transitions.duration.enteringScreen,
    }),
  },
  menuButton: {
    marginLeft: 12,
    marginRight: 36,
  },
  menuButtonHidden: {
    display: 'none',
  },
  title: {
    flexGrow: 1,
  },
  drawerPaper: {
    position: 'relative',
    whiteSpace: 'nowrap',
    width: drawerWidth,
    overflowx: "wrap",
    transition: theme.transitions.create('width', {
      easing: theme.transitions.easing.sharp,
      duration: theme.transitions.duration.enteringScreen,
    }),
  },
  drawerPaperClose: {
    overflowX: 'hidden',
    transition: theme.transitions.create('width', {
      easing: theme.transitions.easing.sharp,
      duration: theme.transitions.duration.leavingScreen,
    }),
    width: 0,
  },
  appBarSpacer: theme.mixins.toolbar,
  content: {
    flexGrow: 1,
    padding: theme.spacing(3),
    height: '100vh',
    overflow: 'auto',
  },
  chartContainer: {
    marginLeft: -22,
  },
  h5: {
    marginBottom: theme.spacing(2),
  },
});

function Dashboard({classes}) {
  const [tagsFilter, setTagsFilter] = React.useState([{key: "", value: ""}]);
  const [stateFilter, setStateFilter] = React.useState("");
  const [pageToken, setPageToken] = React.useState("");
  const [pageSize, setPageSize] = React.useState(25);
  const [nextPageToken, setNextPageToken] = React.useState("");
  const [prevPageToken, setPrevPageToken] = React.useState([]);
  const [open, setOpen] = React.useState(window.innerWidth > 500);

  const handleDrawerOpen = () => {
    setOpen(true);
  };

  const handleDrawerClose = () => {
    setOpen(false);
  };

  //console.log("Dashboard classes:", classes)
  //console.log("Dashboard stateFilter:", stateFilter)
  //console.log("Dashboard tagsFilter:", tagsFilter)

  return (
    <div className={classes.root}>
      <CssBaseline />
      <AppBar
        position="fixed"
        className={classNames(classes.appBar, open && classes.appBarShift)}
      >
        <Toolbar disableGutters={!open} className={classes.toolbar}>
          <IconButton
            color="inherit"
            aria-label="Open drawer"
            onClick={handleDrawerOpen}
            className={classNames(
              classes.menuButton,
              open && classes.menuButtonHidden,
            )}
          >
            <MenuIcon />
          </IconButton>
          <Typography
            component="h1"
            variant="h6"
            color="inherit"
            noWrap
            className={classes.title}
          >
            Funnel
          </Typography>
        </Toolbar>
      </AppBar>
      <Drawer
        variant="permanent"
        classes={{
          paper: classNames(classes.drawerPaper, !open && classes.drawerPaperClose),
        }}
        open={open}
      >
        <div className={classes.toolbarIcon}>
          <IconButton onClick={handleDrawerClose}>
            <ChevronLeftIcon />
          </IconButton>
        </div>
        <Divider />
        <List>{NavListItems}</List>
        <Divider />
        <TaskFilters
          show={window.location.pathname.endsWith("tasks") || window.location.pathname === "/"}
          stateFilter={stateFilter}
          tagsFilter={tagsFilter}
          setStateFilter={setStateFilter}
          setTagsFilter={setTagsFilter}
        />
      </Drawer>
      <main className={classes.content}>
        <div className={classes.appBarSpacer} />
        <Switch>
          <Redirect exact from="/" to="/tasks" />
          <Route exact path="/v1/tasks" render={ () => (
            <TaskList pageSize={pageSize}
                      setPageSize={setPageSize}
                      pageToken={pageToken}
                      setPageToken={setPageToken}
                      nextPageToken={nextPageToken}
                      setNextPageToken={setNextPageToken}
                      prevPageToken={prevPageToken}
                      setPrevPageToken={setPrevPageToken}
                      stateFilter={stateFilter}
                      tagsFilter={tagsFilter} />
          )} />
          <Route exact path="/tasks" render={ () => (
            <TaskList pageSize={pageSize}
                      setPageSize={setPageSize}
                      pageToken={pageToken}
                      setPageToken={setPageToken}
                      nextPageToken={nextPageToken}
                      setNextPageToken={setNextPageToken}
                      prevPageToken={prevPageToken}
                      setPrevPageToken={setPrevPageToken}
                      stateFilter={stateFilter}
                      tagsFilter={tagsFilter} />
          )} />
          <Route exact path="/v1/tasks/:task_id" component={Task} />
          <Route exact path="/tasks/:task_id" component={Task} />
          <Route exact path="/v1/nodes" component={NodeList} />
          <Route exact path="/nodes" component={NodeList} />
          <Route exact path="/v1/nodes/:node_id" component={Node} />
          <Route exact path="/nodes/:node_id" component={Node} />
          <Route exact path="/v1/service-info" component={ServiceInfo} />
          <Route exact path="/service-info" component={ServiceInfo} />
          <Route component={NoMatch} />
        </Switch>
      </main>
    </div>
  );
}

Dashboard.propTypes = {
  classes: PropTypes.object.isRequired,
};

export default withRouter(withStyles(styles)(Dashboard));
