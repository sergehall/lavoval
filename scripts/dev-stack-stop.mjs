import { spawnSync } from 'node:child_process';

const commands = [
  ['pkill', ['-f', 'go run ./cmd/server']],
  ['pkill', ['-f', 'next dev --hostname 127.0.0.1 --port 3000']],
  ['pkill', ['-f', 'node ./scripts/dev-with-open.mjs']]
];

for (const [command, args] of commands) {
  const result = spawnSync(command, args, { stdio: 'inherit' });

  if (result.status !== 0 && result.status !== 1) {
    process.exit(result.status ?? 1);
  }
}

const down = spawnSync('pnpm', ['run', 'infra:down'], { stdio: 'inherit' });
process.exit(down.status ?? 0);
