import { spawn } from 'node:child_process';
import http from 'node:http';

const port = Number(process.env.PORT ?? 3000);
const host = process.env.HOSTNAME ?? '127.0.0.1';
const targetUrl = `http://localhost:${port}`;

const child = spawn('pnpm', ['exec', 'next', 'dev', '--hostname', host, '--port', String(port)], {
  stdio: 'inherit',
  cwd: new URL('..', import.meta.url)
});

let opened = false;

function ping() {
  if (opened) {
    return;
  }

  const request = http.get(`${targetUrl}`, (response) => {
    response.resume();

    if (response.statusCode && response.statusCode < 500) {
      opened = true;
      const opener = spawn('open', [targetUrl], {
        stdio: 'ignore',
        detached: true
      });
      opener.unref();
      return;
    }

    setTimeout(ping, 1000);
  });

  request.on('error', () => {
    setTimeout(ping, 1000);
  });

  request.setTimeout(1000, () => {
    request.destroy();
    setTimeout(ping, 1000);
  });
}

setTimeout(ping, 1200);

child.on('exit', (code, signal) => {
  if (signal) {
    process.kill(process.pid, signal);
    return;
  }

  process.exit(code ?? 0);
});
