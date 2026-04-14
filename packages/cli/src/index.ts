#!/usr/bin/env node

import { spawn } from 'node:child_process';
import { mkdir, readFile, rm, writeFile } from 'node:fs/promises';
import path from 'node:path';
import process from 'node:process';
import { createApiClient, resolveAccessToken, resolveApiBaseUrl } from '../../sdk/src/index.ts';

type ParsedArgs = {
  positionals: string[];
  flags: Record<string, string | boolean>;
};

type CliSession = {
  apiUrl: string;
  accessToken: string;
  refreshToken: string;
  user: {
    id: string;
    email: string;
    role: string;
    firstName?: string;
    lastName?: string;
  };
  storedAt: string;
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
  lavoval auth login --email <email> --password <password> [--api-url URL]
  lavoval auth me
  lavoval auth whoami
  lavoval auth logout
  lavoval skills list [--api-url URL]
  lavoval skills get <skill-id> [--api-url URL]
  lavoval runs list [--token TOKEN] [--api-url URL]
  lavoval runs get <run-id> [--token TOKEN] [--api-url URL]
  lavoval admin runs list [--token TOKEN] [--api-url URL]
  lavoval admin runs get <run-id> [--token TOKEN] [--api-url URL]
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

function sessionFilePath(flags: ParsedArgs['flags']) {
  const explicit = readStringFlag(flags, 'session-file') ?? process.env.LAVOVAL_SESSION_FILE;
  return explicit ? path.resolve(explicit) : path.join(process.cwd(), '.lavoval', 'session.json');
}

async function loadSession(flags: ParsedArgs['flags']) {
  const filePath = sessionFilePath(flags);

  try {
    const raw = await readFile(filePath, 'utf8');
    return JSON.parse(raw) as CliSession;
  } catch (error) {
    if ((error as NodeJS.ErrnoException).code === 'ENOENT') {
      return null;
    }
    throw error;
  }
}

async function saveSession(flags: ParsedArgs['flags'], session: CliSession) {
  const filePath = sessionFilePath(flags);
  await mkdir(path.dirname(filePath), { recursive: true });
  await writeFile(filePath, `${JSON.stringify(session, null, 2)}\n`, 'utf8');
}

async function clearSession(flags: ParsedArgs['flags']) {
  await rm(sessionFilePath(flags), { force: true });
}

async function requireToken(flags: ParsedArgs['flags']) {
  const explicit = resolveAccessToken(readStringFlag(flags, 'token'));
  if (explicit) {
    return explicit;
  }

  const session = await loadSession(flags);
  if (session?.accessToken) {
    return session.accessToken;
  }

  throw new Error(
    'This command requires an access token. Run `lavoval auth login`, pass --token, or set LAVOVAL_ACCESS_TOKEN.',
  );
}

async function resolveClientContext(flags: ParsedArgs['flags']) {
  const session = await loadSession(flags);
  const apiUrl = resolveApiBaseUrl(readStringFlag(flags, 'api-url') ?? session?.apiUrl);
  const client = createApiClient({
    baseUrl: apiUrl,
    fetchFn: fetch,
  });

  return { client, apiUrl, session };
}

function requireFlag(flags: ParsedArgs['flags'], name: string) {
  const value = readStringFlag(flags, name);
  if (!value) {
    throw new Error(`Missing required flag --${name}.`);
  }
  return value;
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

function printSkillDetail(skill: {
  id: string;
  title: string;
  slug: string;
  summary: string;
  provider: string;
  entrypoint: string;
  status: string;
  visibility: string;
  modules: Array<{ title: string; status: string; position: number }>;
}) {
  process.stdout.write(`Skill: ${skill.title}\n`);
  process.stdout.write(`ID: ${skill.id}\n`);
  process.stdout.write(`Slug: ${skill.slug}\n`);
  process.stdout.write(`Status: ${skill.status}\n`);
  process.stdout.write(`Visibility: ${skill.visibility}\n`);
  process.stdout.write(`Runtime: ${skill.provider} -> ${skill.entrypoint}\n`);
  process.stdout.write(`Summary: ${skill.summary}\n`);
  process.stdout.write(`Modules: ${skill.modules.length}\n`);

  if (skill.modules.length > 0) {
    process.stdout.write('Module outline:\n');
    for (const module of skill.modules) {
      process.stdout.write(`  ${module.position + 1}. ${module.title} (${module.status})\n`);
    }
  }
}

function printRunDetail(run: {
  id: string;
  status: string;
  userId: string;
  createdAt: string;
  startedAt?: string | null;
  finishedAt?: string | null;
  errorMessage?: string | null;
  skill: { title: string; slug: string; entrypoint: string };
  meta: { durationMs?: number | null; inputKeysCount: number; outputKeysCount: number; hasError: boolean };
}) {
  process.stdout.write(`Run: ${run.id}\n`);
  process.stdout.write(`Skill: ${run.skill.title} [${run.skill.slug}]\n`);
  process.stdout.write(`Entrypoint: ${run.skill.entrypoint}\n`);
  process.stdout.write(`Status: ${run.status}\n`);
  process.stdout.write(`User ID: ${run.userId}\n`);
  process.stdout.write(`Created: ${run.createdAt}\n`);
  if (run.startedAt) {
    process.stdout.write(`Started: ${run.startedAt}\n`);
  }
  if (run.finishedAt) {
    process.stdout.write(`Finished: ${run.finishedAt}\n`);
  }
  if (run.meta.durationMs) {
    process.stdout.write(`Duration: ${run.meta.durationMs} ms\n`);
  }
  process.stdout.write(`Input fields: ${run.meta.inputKeysCount}\n`);
  process.stdout.write(`Output fields: ${run.meta.outputKeysCount}\n`);
  if (run.meta.hasError && run.errorMessage) {
    process.stdout.write(`Error: ${run.errorMessage}\n`);
  }
}

async function handleSkillsList(parsed: ParsedArgs) {
  const { client } = await resolveClientContext(parsed.flags);
  const response = await client.skills.list();
  printSkillList(response.data);
}

async function handleSkillsGet(parsed: ParsedArgs, skillId: string) {
  const { client } = await resolveClientContext(parsed.flags);
  const response = await client.skills.detail(skillId);
  printSkillDetail(response.data);
}

async function handleRunsList(parsed: ParsedArgs) {
  const { client } = await resolveClientContext(parsed.flags);
  const token = await requireToken(parsed.flags);
  const response = await client.runtime.runs({ token });
  printRunList(response.data);
}

async function handleRunGet(parsed: ParsedArgs, runId: string) {
  const { client } = await resolveClientContext(parsed.flags);
  const token = await requireToken(parsed.flags);
  const response = await client.runtime.runDetail(runId, { token });
  printRunDetail(response.data);
}

async function handleAdminRunsList(parsed: ParsedArgs) {
  const { client } = await resolveClientContext(parsed.flags);
  const token = await requireToken(parsed.flags);
  const response = await client.admin.runs({ token });
  printRunList(response.data);
}

async function handleAdminRunGet(parsed: ParsedArgs, runId: string) {
  const { client } = await resolveClientContext(parsed.flags);
  const token = await requireToken(parsed.flags);
  const response = await client.admin.runDetail(runId, { token });
  printRunDetail(response.data);
}

async function handleRun(parsed: ParsedArgs, skillId: string) {
  const { client } = await resolveClientContext(parsed.flags);
  const token = await requireToken(parsed.flags);
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

async function handleAuthLogin(parsed: ParsedArgs) {
  const { client, apiUrl } = await resolveClientContext(parsed.flags);
  const email = requireFlag(parsed.flags, 'email');
  const password = requireFlag(parsed.flags, 'password');
  const response = await client.auth.login({ email, password });

  await saveSession(parsed.flags, {
    apiUrl,
    accessToken: response.data.accessToken,
    refreshToken: response.data.refreshToken,
    user: response.data.user,
    storedAt: new Date().toISOString(),
  });

  process.stdout.write(
    `Logged in as ${response.data.user.email}. Session saved to ${sessionFilePath(parsed.flags)}\n`,
  );
}

async function handleAuthMe(parsed: ParsedArgs) {
  const { client, session } = await resolveClientContext(parsed.flags);
  const token = await requireToken(parsed.flags);
  const profile = await client.me.profile({ token });

  process.stdout.write(
    `${JSON.stringify(
      {
        session: session?.user ?? null,
        profile: profile.data,
      },
      null,
      2,
    )}\n`,
  );
}

async function handleAuthWhoAmI(parsed: ParsedArgs) {
  const { client, session, apiUrl } = await resolveClientContext(parsed.flags);
  const token = await requireToken(parsed.flags);
  const profile = await client.me.profile({ token });

  process.stdout.write(`Signed in to Lavoval\n`);
  process.stdout.write(`API: ${apiUrl}\n`);
  process.stdout.write(`Email: ${session?.user.email ?? 'unknown'}\n`);
  process.stdout.write(`Role: ${session?.user.role ?? 'unknown'}\n`);
  process.stdout.write(
    `Name: ${profile.data.firstName} ${profile.data.lastName}\n`,
  );
  process.stdout.write(`Timezone: ${profile.data.timezone}\n`);
  process.stdout.write(`Session file: ${sessionFilePath(parsed.flags)}\n`);
  if (session?.storedAt) {
    process.stdout.write(`Stored at: ${session.storedAt}\n`);
  }
}

async function handleAuthLogout(parsed: ParsedArgs) {
  const { client } = await resolveClientContext(parsed.flags);
  const token = await requireToken(parsed.flags);
  await client.auth.logout({ token });
  await clearSession(parsed.flags);
  process.stdout.write('Logged out and cleared local CLI session.\n');
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

  if (command === 'auth' && subcommand === 'login') {
    await handleAuthLogin(parsed);
    return;
  }

  if (command === 'auth' && subcommand === 'me') {
    await handleAuthMe(parsed);
    return;
  }

  if (command === 'auth' && subcommand === 'whoami') {
    await handleAuthWhoAmI(parsed);
    return;
  }

  if (command === 'auth' && subcommand === 'logout') {
    await handleAuthLogout(parsed);
    return;
  }

  if (command === 'skills' && subcommand === 'list') {
    await handleSkillsList(parsed);
    return;
  }

  if (command === 'skills' && subcommand === 'get' && rest[0]) {
    await handleSkillsGet(parsed, rest[0]);
    return;
  }

  if (command === 'runs' && subcommand === 'list') {
    await handleRunsList(parsed);
    return;
  }

  if (command === 'runs' && subcommand === 'get' && rest[0]) {
    await handleRunGet(parsed, rest[0]);
    return;
  }

  if (command === 'admin' && subcommand === 'runs' && rest[0] === 'list') {
    await handleAdminRunsList(parsed);
    return;
  }

  if (command === 'admin' && subcommand === 'runs' && rest[0] === 'get' && rest[1]) {
    await handleAdminRunGet(parsed, rest[1]);
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
