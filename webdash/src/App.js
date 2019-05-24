import React from 'react';
import './App.css';
import Dashboard from './Dashboard.js';
import { BrowserRouter as Router } from "react-router-dom";

class App extends React.Component {

  render() {
    return (
        <Router>
          <Dashboard />
        </Router>
    );
  };
};

export default App;
