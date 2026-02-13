import type { LogEntry } from '../types';

const MAX_ENTRIES = 500;
let logs: LogEntry[] = [];
let logContainer: HTMLElement | null = null;
let autoScroll = true;

export function renderLogViewer(container: HTMLElement) {
  container.innerHTML = `
    <div class="log-viewer">
      <h2>
        <span>Logs</span>
        <button class="btn" id="btn-clear-logs">Clear</button>
      </h2>
      <div class="log-container" id="log-container">
        ${logs.length === 0 ? '<div class="log-empty">No log entries yet. Connect to the VPN to see logs.</div>' : ''}
      </div>
    </div>
  `;

  logContainer = container.querySelector('#log-container');

  // Render existing logs
  if (logs.length > 0) {
    logContainer!.innerHTML = logs.map(formatEntry).join('');
    scrollToBottom();
  }

  // Track auto-scroll: disable if user scrolls up
  logContainer!.addEventListener('scroll', () => {
    const el = logContainer!;
    autoScroll = el.scrollTop + el.clientHeight >= el.scrollHeight - 30;
  });

  container.querySelector('#btn-clear-logs')!.addEventListener('click', () => {
    logs = [];
    if (logContainer) {
      logContainer.innerHTML = '<div class="log-empty">Logs cleared.</div>';
    }
  });
}

export function appendLog(entry: LogEntry) {
  logs.push(entry);
  if (logs.length > MAX_ENTRIES) {
    logs.shift();
  }

  if (logContainer) {
    // Remove empty message if present
    const empty = logContainer.querySelector('.log-empty');
    if (empty) empty.remove();

    logContainer.insertAdjacentHTML('beforeend', formatEntry(entry));

    // Trim DOM nodes to match MAX_ENTRIES
    while (logContainer.children.length > MAX_ENTRIES) {
      logContainer.removeChild(logContainer.firstChild!);
    }

    if (autoScroll) {
      scrollToBottom();
    }
  }
}

function scrollToBottom() {
  if (logContainer) {
    logContainer.scrollTop = logContainer.scrollHeight;
  }
}

function formatEntry(entry: LogEntry): string {
  return `<div class="log-entry"><span class="log-time">${escapeHtml(entry.timestamp)}</span><span class="log-msg">${escapeHtml(entry.message)}</span></div>`;
}

function escapeHtml(text: string): string {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}
