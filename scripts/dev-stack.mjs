import { spawn, spawnSync } from 'node:child_process';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const rootDir = path.resolve(__dirname, '..');

function run(name, command, args, options = {}) {
  const child = spawn(command, args, {
    cwd: rootDir,
    stdio: 'pipe',
    env: process.env,
    ...options
  });

  child.stdout.on('data', (chunk) => {
    process.stdout.write(`[${name}] ${chunk}`);
  });

  child.stderr.on('data', (chunk) => {
    process.stderr.write(`[${name}] ${chunk}`);
  });

  child.on('exit', (code, signal) => {
    if (signal) {
      process.stderr.write(`[${name}] exited via signal ${signal}\n`);
      return;
    }

    if (code && code !== 0) {
      process.stderr.write(`[${name}] exited with code ${code}\n`);
      process.exitCode = code;
    }
  });

  return child;
}

const infra = spawn('pnpm', ['run', 'infra:up'], {
  cwd: rootDir,
  stdio: 'inherit',
  env: process.env
});

infra.on('exit', (code) => {
  if (code !== 0) {
    process.exit(code ?? 1);
    return;
  }

  const migrate = spawnSync('pnpm', ['run', 'db:migrate'], {
    cwd: rootDir,
    stdio: 'inherit',
    env: process.env
  });

  if (migrate.status !== 0) {
    process.exit(migrate.status ?? 1);
    return;
  }

  const api = run('api', 'go', ['run', './cmd/server'], {
    cwd: path.join(rootDir, 'apps/api')
  });

  const web = run('web', 'pnpm', ['run', 'dev:web']);

  const children = [api, web];

  const shutdown = () => {
    for (const child of children) {
      if (!child.killed) {
        child.kill('SIGINT');
      }
    }
  };

  process.on('SIGINT', shutdown);
  process.on('SIGTERM', shutdown);
});
