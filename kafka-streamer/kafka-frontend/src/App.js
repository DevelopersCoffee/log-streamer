import React, { useEffect, useState, useRef } from 'react';
import { JSONTree } from 'react-json-tree';
import moment from 'moment';
import './App.css';

function App() {
  const [logs, setLogs] = useState([]);
  const [topics, setTopics] = useState([]);
  const [selectedTopic, setSelectedTopic] = useState('test-topic');
  const [messageLimit, setMessageLimit] = useState(50);
  const [autoScroll, setAutoScroll] = useState(true);
  const logContainerRef = useRef(null);

  useEffect(() => {
    fetch('/api/topics')
      .then(response => response.json())
      .then(data => setTopics(data.topics))
      .catch(error => console.error('Error fetching topics:', error));
  }, []);

  useEffect(() => {
    if (!selectedTopic) return;

    const eventSource = new EventSource(`/api/events/${selectedTopic}/${messageLimit}`);

    eventSource.onmessage = function(event) {
      setLogs((prevLogs) => [...prevLogs.slice(-messageLimit), event.data]);
    };

    eventSource.onerror = function(event) {
      console.error('EventSource failed:', event);
      eventSource.close();
    };

    return () => {
      eventSource.close();
    };
  }, [selectedTopic, messageLimit]);

  useEffect(() => {
    if (autoScroll && logContainerRef.current) {
      logContainerRef.current.scrollTop = logContainerRef.current.scrollHeight;
    }
  }, [logs, autoScroll]);

  const theme = {
    base00: '#1e1e1e',
    base01: '#282c34',
    base02: '#383c4a',
    base03: '#4b5263',
    base04: '#6272a4',
    base05: '#e9efff',
    base06: '#b45bcf',
    base07: '#f1fa8c',
    base08: '#ff79c6',
    base09: '#ffb86c',
    base0A: '#bd93f9',
    base0B: '#50fa7b',
    base0C: '#8be9fd',
    base0D: '#6272a4',
    base0E: '#ff79c6',
    base0F: '#f8f8f2'
  };

  const formatTimestamp = (timestamp) => {
    return moment(timestamp).format('YYYY-MM-DD HH:mm:ss');
  };

  const renderLog = (log) => {
    try {
      const jsonLog = JSON.parse(log);
      if (jsonLog.time) {
        jsonLog.time = formatTimestamp(jsonLog.time);
      }
      return <JSONTree data={jsonLog} theme={theme} invertTheme={false} hideRoot />;
    } catch (e) {
      return <div className="plain-log">{log}</div>;
    }
  };

  return (
    <div className="App">
      <header className="App-header">
        <h1>Kafka Message Stream</h1>
        <div className="controls">
          <select value={selectedTopic} onChange={e => setSelectedTopic(e.target.value)}>
            {topics.map(topic => (
              <option key={topic} value={topic}>{topic}</option>
            ))}
          </select>
          <input
            type="number"
            value={messageLimit}
            onChange={e => setMessageLimit(Number(e.target.value))}
            min="1"
          />
          <button onClick={() => setAutoScroll(!autoScroll)}>
            {autoScroll ? 'Disable Auto-scroll' : 'Enable Auto-scroll'}
          </button>
        </div>
        <div className="logs-container" ref={logContainerRef}>
          {logs.map((log, index) => (
            <div key={index} className="log-entry">
              {renderLog(log)}
            </div>
          ))}
        </div>
      </header>
    </div>
  );
}

export default App;
