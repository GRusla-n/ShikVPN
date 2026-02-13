import type { Page, StatusUpdate, LogEntry } from './types';
import { renderSidebar } from './components/Sidebar';
import { renderConnectionPanel, updateConnectionStatus } from './components/ConnectionPanel';
import { renderConfigEditor } from './components/ConfigEditor';
import { renderLogViewer, appendLog } from './components/LogViewer';

// Wails runtime events
declare global {
  interface Window {
    runtime: {
      EventsOn(event: string, callback: (...args: any[]) => void): void;
    };
  }
}

let currentPage: Page = 'connection';

function navigate(page: Page) {
  currentPage = page;
  const sidebar = document.getElementById('sidebar')!;
  const content = document.getElementById('content')!;

  renderSidebar(sidebar, currentPage, navigate);
  renderPage(content);
}

function renderPage(content: HTMLElement) {
  switch (currentPage) {
    case 'connection':
      renderConnectionPanel(content);
      break;
    case 'config':
      renderConfigEditor(content);
      break;
    case 'logs':
      renderLogViewer(content);
      break;
  }
}

// Wire up Wails events
function setupEvents() {
  window.runtime.EventsOn('vpn:status', (status: StatusUpdate) => {
    if (currentPage === 'connection') {
      const content = document.getElementById('content')!;
      updateConnectionStatus(status, content);
    }
  });

  window.runtime.EventsOn('vpn:log', (entry: LogEntry) => {
    appendLog(entry);
  });
}

// Initialize on DOM ready
document.addEventListener('DOMContentLoaded', () => {
  navigate('connection');
  setupEvents();
});
