import '@xterm/xterm/css/xterm.css';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import React, { useRef, useEffect } from 'react';

const CLIView: React.FC = () => {
  const terminalRef = useRef<HTMLDivElement>(null);
  const terminal = useRef<Terminal>();
  const fitAddon = useRef(new FitAddon());

  useEffect(() => {
    if (terminalRef.current) {
      terminal.current = new Terminal({
        cursorBlink: true,
        theme: {
          background: '#000000',
        },
      });
      terminal.current.loadAddon(fitAddon.current);
      terminal.current.open(terminalRef.current);
      fitAddon.current.fit();
    }
  }, []);

  return <div ref={terminalRef} style={{ height: '500px', width: '100%' }} />;
};

export default CLIView;
