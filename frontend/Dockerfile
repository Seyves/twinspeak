# Build stage
FROM node:24-alpine AS builder

WORKDIR /app

ENV NODE_ENV=production

RUN corepack enable && corepack prepare pnpm@11.5.2 --activate

COPY package.json pnpm-lock.yaml ./

RUN pnpm install --frozen-lockfile

COPY . .

RUN pnpm run build

# https://github.com/TanStack/router/issues/7034
# Fix: Tailwind CSS v4 generates different CSS for SSR vs client builds
# The SSR build includes extra utilities from node_modules/shadcn that aren't used
# Solution: Replace SSR CSS with client CSS to ensure consistency
RUN CLIENT_CSS=$(find .output/public/assets -name "styles-*.css" -exec basename {} \; | head -1) && \
    SSR_CSS=$(find node_modules/.nitro/vite/services/ssr/assets -name "styles-*.css" -exec basename {} \; 2>/dev/null | head -1) && \
    if [ -n "$CLIENT_CSS" ] && [ -n "$SSR_CSS" ] && [ "$CLIENT_CSS" != "$SSR_CSS" ]; then \
        echo "=== Fixing CSS Mismatch ===" && \
        echo "Client CSS: $CLIENT_CSS" && \
        echo "SSR CSS: $SSR_CSS" && \
        echo "Replacing SSR CSS references with client CSS..." && \
        find .output/server/_ssr -name "*.mjs" -exec sed -i "s|$SSR_CSS|$CLIENT_CSS|g" {} \; && \
        echo "✓ CSS references unified"; \
    elif [ "$CLIENT_CSS" = "$SSR_CSS" ]; then \
        echo "✓ CSS already matches: $CLIENT_CSS"; \
    else \
        echo "⚠ Warning: Could not find CSS files to compare"; \
    fi

FROM node:24-alpine AS runtime

WORKDIR /app

RUN corepack enable && corepack prepare pnpm@11.5.2 --activate

COPY package.json pnpm-lock.yaml ./

RUN pnpm install --prod --frozen-lockfile

# Create non-root user
# RUN addgroup -g 1001 -S nodejs && \
#     adduser -S nodejs -u 1001

# Copy built application from builder with correct ownership
COPY --from=builder /app/.output ./.output

# USER nodejs

EXPOSE 4321

ENV NODE_ENV=production
ENV PORT=4321

CMD ["node", ".output/server/index.mjs"]
