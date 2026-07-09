FROM node:22-alpine

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

RUN chown -R appuser:appgroup /app
USER appuser

EXPOSE 3000
HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD wget -qO- http://localhost:3000/en/health || exit 1
STOPSIGNAL SIGTERM
CMD ["npm", "start"]
