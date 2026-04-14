FROM node:22-alpine AS deps
WORKDIR /app
COPY package.json pnpm-workspace.yaml ./
COPY apps/web/package.json ./apps/web/package.json
COPY packages/contracts/package.json ./packages/contracts/package.json
RUN corepack enable && pnpm install --filter @lavoval/web... --filter @lavoval/contracts...

FROM node:22-alpine AS builder
WORKDIR /app
COPY --from=deps /app /app
COPY . .
RUN corepack enable && pnpm --dir apps/web build

FROM node:22-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
COPY --from=builder /app .
EXPOSE 3000
CMD ["sh", "-c", "corepack enable && pnpm --dir apps/web start"]
