import React from 'react';
import PropTypes from 'prop-types';

import InertiaClient from '../client';

const SidebarHeader = ({ children }) => (
  <div style={sidebarHeaderStyles.container}>
    <a href="/#/home" onClick={() => { return false; }} style={sidebarHeaderStyles.text}>
      {children}
    </a>
  </div>
);
const sidebarHeaderStyles = {
  container: {
    display: 'flex',
    alignItems: 'center',
    height: '3rem',
    width: '100%',
    paddingLeft: '2rem'
  },

  text: {
    textDecoration: 'none',
    color: '#5f5f5f'
  }
};

const SidebarButton = ({ children }) => (
  <div style={sidebarButtonStyles.container}>
    <a href="/#/home" onClick={() => { }} style={sidebarButtonStyles.text}>
      {children}
    </a>
  </div>
);
const sidebarButtonStyles = {
  container: {
    display: 'flex',
    alignItems: 'center',
    height: '3rem',
    width: '100%',
    paddingLeft: '3rem'
  },

  text: {
    textDecoration: 'none',
    color: '#101010'
  }
};

const SidebarText = ({ children }) => (
  <div style={sidebarButtonStyles.container}>
    <p style={sidebarButtonStyles.text}>
      {children}
    </p>
  </div>
);
const sidebarTextStyles = {
  container: {
    display: 'flex',
    alignItems: 'center',
    height: '3rem',
    width: '100%',
    paddingLeft: '3rem'
  },

  text: {
    textDecoration: 'none',
    color: '#101010'
  }
};

export default class Home extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      loading: true,

      remoteVersion: '',

      repoBranch: '',
      repoCommitHash: '',
      repoCommitMessage: '',
      repoBuildType: '',
      repoBuilding: false,
      containers: [],
    };

    this.handleLogout = this.handleLogout.bind(this);
    this.handleGetStatus = this.handleGetStatus.bind(this);

    this.handleGetStatus()
      .then(() => this.setState({ loading: false }))
      .catch((err) => console.error(err));
  }

  async handleLogout() {
    const response = await this.props.client.logout();
    if (response.status != 200) console.error(response);
    this.props.history.push('/login');
  }

  async handleGetStatus() {
    const response = await this.props.client.getRemoteStatus();
    if (response.status !== 200) return new Error('bad response: ' + response);
    const status = await response.json();
    this.setState({
      remoteVersion: status.version,
      repoBranch: status.branch,
      repoBuilding: status.build_active,
      repoCommitHash: status.commit_hash,
      repoCommitMessage: status.commit_message,
      containers: status.containers,
    });
  }

  render() {
    const containers = this.state.containers.map((c) =>
      <SidebarButton>{c}</SidebarButton>
    );

    const buildMessage = this.state.repoBuilding
      ? <SidebarText>{buildMessage}</SidebarText>
      : <SidebarText>No deployment active</SidebarText>;
    const repoState = this.state.repoCommitHash
      ? (
        <div>
          <SidebarText>{this.state.repoBranch} ({this.state.repoBuildType})</SidebarText>
          <SidebarText>{this.state.repoCommitMessage} ({this.state.repoCommitHash})</SidebarText>
        </div>
      )
      : null;

    return (
      <div style={styles.container}>

        <header style={styles.headerBar}>
          <p style={{ fontWeight: 500, fontSize: 24, color: '#101010' }}>Inertia Web</p>
          <p align='center' style={{ fontWeight: 300, fontSize: 12, color: '#101010' }}>{this.state.remoteVersion}</p>
          <a onClick={this.handleLogout} style={{ textDecoration: 'none', color: '#5f5f5f' }}>logout</a>
        </header>

        <div style={styles.innerContainer}>

          <div style={styles.sidebar}>
            <SidebarHeader>Status</SidebarHeader>
            {buildMessage}
            {repoState}
            <SidebarHeader>Containers</SidebarHeader>
            {containers}
          </div>

          <div style={styles.main}>
            <div style={{ display: 'flex', height: '100%', alignItems: 'center', justifyContent: 'center' }}>
              <p style={styles.underConstruction}>coming soon!</p>
            </div>
          </div>

        </div>
      </div>
    );
  }
}

Home.propTypes = {
  client: PropTypes.instanceOf(InertiaClient),
};

// hardcode all styles for now, until we flesh out UI
const styles = {
  container: {
    display: 'flex',
    flexFlow: 'column',
    height: '100%',
    width: '100%'
  },

  innerContainer: {
    display: 'flex',
    flexFlow: 'row',
    height: '100%',
    width: '100%'
  },

  headerBar: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    width: '100%',
    height: '4rem',
    padding: '0 2rem',
    borderBottom: '1px solid #c1c1c1'
  },

  sidebar: {
    display: 'flex',
    flexFlow: 'column',
    width: '20rem',
    height: '100%',
    paddingTop: '0.5rem',
    borderRight: '1px solid #c1c1c1',
    backgroundColor: '#f0f0f0'
  },

  main: {
    height: '100%',
    width: '100%',
    overflowY: 'scroll'
  },

  button: {
    flex: 'none'
  },

  underConstruction: {
    textAlign: 'center',
    fontSize: 24,
    color: '#9f9f9f'
  }
};
