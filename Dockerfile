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

# Build the application
RUN npm run build

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
COPY --from=build /app/php.ini ./php.ini
COPY --from=build /app/package.json ./package.json

# Install only production dependencies
RUN npm install --only=production

# Expose port
EXPOSE 4173

# Start the application
CMD ["npm", "run", "preview"]