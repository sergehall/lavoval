#!/usr/bin/env node

import { spawn } from 'node:child_process';
import process from 'node:process';
import { createApiClient, resolveAccessToken, resolveApiBaseUrl } from '../../sdk/src/index.ts';

type ParsedArgs = {
  positionals: string[];
  flags: Record<string, string | boolean>;
};

function parseArgs(argv: string[]): ParsedArgs {
  const positionals: string[] = [];
  const flags: Record<string, string | boolean> = {};

  for (let index = 0; index < argv.length; index += 1) {
    const token = argv[index];
    if (!token.startsWith('--')) {
      positionals.push(token);
      continue;
    }

    const key = token.slice(2);
    const next = argv[index + 1];
    if (!next || next.startsWith('--')) {
      flags[key] = true;
      continue;
    }

    flags[key] = next;
    index += 1;
  }

  return { positionals, flags };
}

function printHelp() {
  process.stdout.write(`Lavoval CLI

Usage:
  lavoval dev
  lavoval skills list [--api-url URL]
  lavoval runs list [--token TOKEN] [--api-url URL]
  lavoval run <skill-id> [--text TEXT] [--token TOKEN] [--api-url URL]
  lavoval skill run <skill-id> [--text TEXT] [--token TOKEN] [--api-url URL]

Environment:
  LAVOVAL_API_URL       Base API URL. Defaults to http://localhost:8080
  LAVOVAL_ACCESS_TOKEN  Access token for authenticated commands
`);
}

function readStringFlag(flags: ParsedArgs['flags'], name: string) {
  const value = flags[name];
  return typeof value === 'string' ? value : undefined;
}

function requireToken(flags: ParsedArgs['flags']) {
  const token = resolveAccessToken(readStringFlag(flags, 'token'));
  if (!token) {
    throw new Error(
      'This command requires an access token. Pass --token or set LAVOVAL_ACCESS_TOKEN.',
    );
  }
  return token;
}

function createClient(flags: ParsedArgs['flags']) {
  return createApiClient({
    baseUrl: resolveApiBaseUrl(readStringFlag(flags, 'api-url')),
    fetchFn: fetch,
  });
}

function printSkillList(skills: Array<{ id: string; title: string; slug: string; entrypoint: string; status: string }>) {
  if (skills.length === 0) {
    process.stdout.write('No skills found.\n');
    return;
  }

  for (const skill of skills) {
    process.stdout.write(
      `${skill.id}  ${skill.title}  [${skill.slug}]  ${skill.entrypoint}  ${skill.status}\n`,
    );
  }
}

function printRunList(runs: Array<{ id: string; status: string; createdAt: string; skill: { title: string; entrypoint: string } }>) {
  if (runs.length === 0) {
    process.stdout.write('No runs found.\n');
    return;
  }

  for (const run of runs) {
    process.stdout.write(
      `${run.id}  ${run.skill.title}  ${run.skill.entrypoint}  ${run.status}  ${run.createdAt}\n`,
    );
  }
}

async function handleSkillsList(parsed: ParsedArgs) {
  const client = createClient(parsed.flags);
  const response = await client.skills.list();
  printSkillList(response.data);
}

async function handleRunsList(parsed: ParsedArgs) {
  const client = createClient(parsed.flags);
  const token = requireToken(parsed.flags);
  const response = await client.runtime.runs({ token });
  printRunList(response.data);
}

async function handleRun(parsed: ParsedArgs, skillId: string) {
  const client = createClient(parsed.flags);
  const token = requireToken(parsed.flags);
  const text = readStringFlag(parsed.flags, 'text');
  const response = await client.runtime.run(
    {
      skillId,
      input: text ? { text } : {},
    },
    { token },
  );

  process.stdout.write(`${JSON.stringify(response.data, null, 2)}\n`);
}

function runDevCommand() {
  const child = spawn('pnpm', ['run', 'dev:stack'], {
    cwd: process.cwd(),
    stdio: 'inherit',
    env: process.env,
  });

  child.on('exit', (code, signal) => {
    if (signal) {
      process.kill(process.pid, signal);
      return;
    }

    process.exit(code ?? 0);
  });
}

async function main() {
  const parsed = parseArgs(process.argv.slice(2));
  const [command, subcommand, ...rest] = parsed.positionals;

  if (!command || command === 'help' || command === '--help') {
    printHelp();
    return;
  }

  if (command === 'dev') {
    runDevCommand();
    return;
  }

  if (command === 'skills' && subcommand === 'list') {
    await handleSkillsList(parsed);
    return;
  }

  if (command === 'runs' && subcommand === 'list') {
    await handleRunsList(parsed);
    return;
  }

  if (command === 'run' && subcommand) {
    await handleRun(parsed, subcommand);
    return;
  }

  if (command === 'skill' && subcommand === 'run' && rest[0]) {
    await handleRun(parsed, rest[0]);
    return;
  }

  throw new Error('Unknown command. Run `lavoval help` to see available commands.');
}

main().catch((error) => {
  const message =
    error instanceof Error && error.message === 'fetch failed'
      ? 'Could not reach the Lavoval API. Start the stack with `lavoval dev` or set --api-url / LAVOVAL_API_URL.'
      : error instanceof Error
        ? error.message
        : String(error);
  process.stderr.write(`lavoval: ${message}\n`);
  process.exit(1);
});
