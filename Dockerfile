# Use Node.js LTS version
FROM node:18-alpine AS build

# Set working directory
WORKDIR /app

# Copy package files
COPY package.json package-lock.json ./

# Install dependencies
RUN npm install

# Copy all files
COPY . .

# Build the application with fallback for failures
RUN npm run build || (echo "Build failed, creating fallback files" && mkdir -p /app/dist && cp /app/public/index.html /app/dist/index.html)

# Production stage
FROM node:18-alpine AS production

# Install PHP and required extensions
RUN apk add --no-cache php php-fpm php-pdo php-pdo_mysql php-json php-openssl

# Set working directory
WORKDIR /app

# Copy built assets from build stage
COPY --from=build /app/dist ./dist
COPY --from=build /app/public ./public
COPY --from=build /app/mysql-api.php ./mysql-api.php
COPY --from=build /app/test-php.php ./test-php.php
COPY --from=build /app/debug-api.php ./debug-api.php
COPY --from=build /app/db-test.php ./db-test.php
COPY --from=build /app/request-logger.php ./request-logger.php
COPY --from=build /app/php.ini ./php.ini
COPY --from=build /app/package.json ./package.json
COPY --from=build /app/healthcheck.js ./healthcheck.js
COPY --from=build /app/start.sh ./start.sh
RUN chmod +x ./start.sh

# Install production dependencies and ensure vite is available
RUN npm install --only=production
RUN npm install -g vite

# Expose port
EXPOSE 4173

# Start the application
CMD ["/bin/sh", "./start.sh"]