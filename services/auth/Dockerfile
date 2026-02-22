# Stage 1: Build the application
FROM node:24-alpine AS builder

WORKDIR /app

# Copy package files first to leverage Docker cache for dependencies
COPY package*.json ./

# Install all dependencies (including devDependencies for building)
RUN npm ci

# Copy the rest of the application source code
COPY . .

# Build the application
RUN npm run build

# Stage 2: Run the application
FROM node:24-alpine AS runner

WORKDIR /app

ENV NODE_ENV=production

# Copy package files again for production dependencies
COPY package*.json ./

# Install only production dependencies
RUN npm ci --only=production && npm cache clean --force

# Copy the built application from the builder stage
COPY --from=builder /app/dist ./dist

# Expose the application port (defaults to 8000 based on validation schema)
EXPOSE 8001

# Start the application
CMD ["node", "dist/src/main"]
